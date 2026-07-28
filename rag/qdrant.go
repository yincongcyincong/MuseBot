package rag

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	qdrant "github.com/qdrant/go-client/qdrant"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/langchaingo/embeddings"
	"github.com/yincongcyincong/langchaingo/schema"
	"github.com/yincongcyincong/langchaingo/vectorstores"
)

const (
	defaultQdrantCollection = "MuseBot"
	qdrantContentKey        = "content"
)

type qdrantClient interface {
	CollectionExists(context.Context, string) (bool, error)
	CreateCollection(context.Context, *qdrant.CreateCollection) error
	Upsert(context.Context, *qdrant.UpsertPoints) (*qdrant.UpdateResult, error)
	Query(context.Context, *qdrant.QueryPoints) ([]*qdrant.ScoredPoint, error)
	Delete(context.Context, *qdrant.DeletePoints) (*qdrant.UpdateResult, error)
	Close() error
}

type QdrantStore struct {
	client         qdrantClient
	embedder       embeddings.Embedder
	collectionName string
}

var _ vectorstores.VectorStore = (*QdrantStore)(nil)

func NewQdrantStore(ctx context.Context, embedder embeddings.Embedder) (*QdrantStore, error) {
	if embedder == nil {
		return nil, errors.New("qdrant embedder is not configured")
	}

	clientConfig, collectionName, err := qdrantSettings()
	if err != nil {
		return nil, err
	}

	client, err := qdrant.NewClient(clientConfig)
	if err != nil {
		return nil, fmt.Errorf("create qdrant client: %w", err)
	}

	store := &QdrantStore{
		client:         client,
		embedder:       embedder,
		collectionName: collectionName,
	}
	if err = store.ensureCollection(ctx); err != nil {
		_ = client.Close()
		return nil, err
	}
	return store, nil
}

func qdrantSettings() (*qdrant.Config, string, error) {
	host := strings.TrimSpace(conf.RagConfInfo.QdrantHost)
	if host == "" {
		host = "localhost"
	}
	if strings.Contains(host, "://") {
		return nil, "", errors.New("qdrant host must not include a URL scheme")
	}

	port := conf.RagConfInfo.QdrantPort
	if port == 0 {
		port = 6334
	}

	collectionName := strings.TrimSpace(conf.RagConfInfo.Space)
	if collectionName == "" {
		collectionName = defaultQdrantCollection
	}

	return &qdrant.Config{
		Host:                   host,
		Port:                   port,
		APIKey:                 conf.RagConfInfo.QdrantAPIKey,
		UseTLS:                 conf.RagConfInfo.QdrantUseTLS,
		PoolSize:               1,
		SkipCompatibilityCheck: true,
	}, collectionName, nil
}

func (s *QdrantStore) ensureCollection(ctx context.Context) error {
	exists, err := s.client.CollectionExists(ctx, s.collectionName)
	if err != nil {
		return fmt.Errorf("check qdrant collection %q: %w", s.collectionName, err)
	}
	if exists {
		return nil
	}

	vector, err := s.embedder.EmbedQuery(ctx, "MuseBot collection initialization")
	if err != nil {
		return fmt.Errorf("detect qdrant vector size: %w", err)
	}
	if len(vector) == 0 {
		return errors.New("detect qdrant vector size: embedder returned an empty vector")
	}

	err = s.client.CreateCollection(ctx, &qdrant.CreateCollection{
		CollectionName: s.collectionName,
		VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
			Size:     uint64(len(vector)),
			Distance: qdrant.Distance_Cosine,
		}),
	})
	if err == nil {
		return nil
	}

	return fmt.Errorf("create qdrant collection %q: %w", s.collectionName, err)
}

func (s *QdrantStore) AddDocuments(
	ctx context.Context,
	docs []schema.Document,
	options ...vectorstores.Option,
) ([]string, error) {
	if len(docs) == 0 {
		return []string{}, nil
	}

	opts := applyVectorStoreOptions(options)
	embedder := s.embedder
	if opts.Embedder != nil {
		embedder = opts.Embedder
	}

	texts := make([]string, len(docs))
	for i := range docs {
		texts[i] = docs[i].PageContent
	}
	vectors, err := embedder.EmbedDocuments(ctx, texts)
	if err != nil {
		return nil, fmt.Errorf("embed qdrant documents: %w", err)
	}
	if len(vectors) != len(docs) {
		return nil, fmt.Errorf("embedder returned %d vectors for %d documents", len(vectors), len(docs))
	}

	ids := make([]string, len(docs))
	points := make([]*qdrant.PointStruct, len(docs))
	for i := range docs {
		if len(vectors[i]) == 0 {
			return nil, fmt.Errorf("embedder returned an empty vector for document %d", i)
		}

		metadata := make(map[string]interface{}, len(docs[i].Metadata)+1)
		for key, value := range docs[i].Metadata {
			stringValue, ok := value.(string)
			if !ok {
				return nil, fmt.Errorf(
					"metadata %q for document %d must be a string, got %T",
					key,
					i,
					value,
				)
			}
			metadata[key] = stringValue
		}
		metadata[qdrantContentKey] = docs[i].PageContent

		payload, payloadErr := qdrant.TryValueMap(metadata)
		if payloadErr != nil {
			return nil, fmt.Errorf("convert metadata for document %d: %w", i, payloadErr)
		}

		ids[i] = uuid.NewString()
		points[i] = &qdrant.PointStruct{
			Id:      qdrant.NewIDUUID(ids[i]),
			Vectors: qdrant.NewVectorsDense(vectors[i]),
			Payload: payload,
		}
	}

	_, err = s.client.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: s.collectionName,
		Wait:           qdrant.PtrOf(true),
		Points:         points,
	})
	if err != nil {
		return nil, fmt.Errorf("upsert qdrant points: %w", err)
	}
	return ids, nil
}

func (s *QdrantStore) SimilaritySearch(
	ctx context.Context,
	query string,
	numDocuments int,
	options ...vectorstores.Option,
) ([]schema.Document, error) {
	if numDocuments <= 0 {
		return nil, errors.New("number of documents must be greater than zero")
	}

	opts := applyVectorStoreOptions(options)
	if opts.ScoreThreshold < 0 || opts.ScoreThreshold > 1 {
		return nil, errors.New("score threshold must be between 0 and 1")
	}

	var filter *qdrant.Filter
	switch value := opts.Filters.(type) {
	case nil:
	case *qdrant.Filter:
		filter = value
	default:
		return nil, fmt.Errorf("qdrant filters must be *qdrant.Filter, got %T", opts.Filters)
	}

	embedder := s.embedder
	if opts.Embedder != nil {
		embedder = opts.Embedder
	}
	vector, err := embedder.EmbedQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("embed qdrant query: %w", err)
	}
	if len(vector) == 0 {
		return nil, errors.New("embed qdrant query: embedder returned an empty vector")
	}

	request := &qdrant.QueryPoints{
		CollectionName: s.collectionName,
		Query:          qdrant.NewQueryDense(vector),
		Filter:         filter,
		Limit:          qdrant.PtrOf(uint64(numDocuments)),
		WithPayload:    qdrant.NewWithPayload(true),
	}
	if opts.ScoreThreshold != 0 {
		request.ScoreThreshold = qdrant.PtrOf(opts.ScoreThreshold)
	}

	points, err := s.client.Query(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("query qdrant points: %w", err)
	}

	documents := make([]schema.Document, 0, len(points))
	for i, point := range points {
		if point == nil {
			return nil, fmt.Errorf("qdrant returned a nil point at index %d", i)
		}
		metadata := make(map[string]interface{}, len(point.Payload)-1)
		content := ""
		contentFound := false
		for key, value := range point.Payload {
			if value == nil {
				return nil, fmt.Errorf("qdrant result %d has a nil %q payload", i, key)
			}
			stringValue, ok := value.Kind.(*qdrant.Value_StringValue)
			if !ok {
				return nil, fmt.Errorf("qdrant result %d payload %q is not a string", i, key)
			}
			if key == qdrantContentKey {
				content = stringValue.StringValue
				contentFound = true
				continue
			}
			metadata[key] = stringValue.StringValue
		}
		if !contentFound {
			return nil, fmt.Errorf("qdrant result %d has no string %q payload", i, qdrantContentKey)
		}
		documents = append(documents, schema.Document{
			PageContent: content,
			Metadata:    metadata,
			Score:       point.Score,
		})
	}
	return documents, nil
}

func (s *QdrantStore) Delete(ctx context.Context, ids []string) error {
	pointIDs := make([]*qdrant.PointId, 0, len(ids))
	seen := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		if _, err := uuid.Parse(id); err != nil {
			return fmt.Errorf("invalid qdrant point ID %q: %w", id, err)
		}
		seen[id] = struct{}{}
		pointIDs = append(pointIDs, qdrant.NewIDUUID(id))
	}
	if len(pointIDs) == 0 {
		return nil
	}

	_, err := s.client.Delete(ctx, &qdrant.DeletePoints{
		CollectionName: s.collectionName,
		Wait:           qdrant.PtrOf(true),
		Points:         qdrant.NewPointsSelector(pointIDs...),
	})
	if err != nil {
		return fmt.Errorf("delete qdrant points: %w", err)
	}
	return nil
}

func (s *QdrantStore) Close() error {
	if s == nil || s.client == nil {
		return nil
	}
	return s.client.Close()
}

func applyVectorStoreOptions(options []vectorstores.Option) vectorstores.Options {
	opts := vectorstores.Options{}
	for _, option := range options {
		if option != nil {
			option(&opts)
		}
	}
	return opts
}
