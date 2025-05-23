package conf

import (
	"context"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yincongcyincong/langchaingo/documentloaders"
	"github.com/yincongcyincong/langchaingo/embeddings"
	"github.com/yincongcyincong/langchaingo/llms/ernie"
	"github.com/yincongcyincong/langchaingo/llms/googleai"
	"github.com/yincongcyincong/langchaingo/llms/openai"
	"github.com/yincongcyincong/langchaingo/schema"
	"github.com/yincongcyincong/langchaingo/textsplitter"
	"github.com/yincongcyincong/langchaingo/vectorstores"
	"github.com/yincongcyincong/langchaingo/vectorstores/chroma"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
)

var (
	EmbeddingType *string
	KnowledgePath *string
	VectorDBType  *string

	ChromaURL   *string
	ChromaSpace *string
)

func InitRagConf() {
	EmbeddingType = flag.String("embedding_type", "", "embedding split api: openai gemini ernie")
	KnowledgePath = flag.String("knowledge_path", "./data/knowledge", "knowledge")
	VectorDBType = flag.String("vector_db_type", "chroma", "vector db type: chroma")

	ChromaURL = flag.String("chroma_url", "http://localhost:8000", "chroma url")
	ChromaSpace = flag.String("chroma_space", "telegram-deepseek-bot", "chroma space")

	if os.Getenv("EMBEDDING_TYPE") != "" {
		*EmbeddingType = os.Getenv("EMBEDDING_TYPE")
	}

	if os.Getenv("KNOWLEDGE_PATH") != "" {
		*KnowledgePath = os.Getenv("KNOWLEDGE_PATH")
	}

	if os.Getenv("VECTOR_DB_TYPE") != "" {
		*VectorDBType = os.Getenv("VECTOR_DB_TYPE")
	}

	if os.Getenv("CHROMA_URL") != "" {
		*ChromaURL = os.Getenv("CHROMA_URL")
	}

	if os.Getenv("CHROMA_SPACE") != "" {
		*ChromaSpace = os.Getenv("CHROMA_SPACE")
	}

	logger.Info("RAG_CONF", "EmbeddingType", *EmbeddingType)
	logger.Info("RAG_CONF", "KnowledgePath", *KnowledgePath)
	logger.Info("RAG_CONF", "VectorDBType", *VectorDBType)
	logger.Info("RAG_CONF", "ChromaURL", *ChromaURL)
	logger.Info("RAG_CONF", "ChromaSpace", *ChromaSpace)
}

func InitRag() {
	if *EmbeddingType == "" {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	var embedder embeddings.Embedder
	var err error
	switch *EmbeddingType {
	case "openai":
		embedder, err = initOpenAIEmbedding()
	case "gemini":
		embedder, err = initGeminiEmbedding(ctx)
	case "ernie":
		embedder, err = initErnieEmbedding()
	default:
		logger.Error("embedding type not exist", "embedding type", *EmbeddingType)
		return
	}

	if err != nil {
		logger.Error("init embedding fail", "err", err)
		return
	}

	var store vectorstores.VectorStore
	switch *VectorDBType {
	case "chroma":
		store, err = chroma.NewV2(
			chroma.WithChromaURLV2("http://localhost:8000"),
			chroma.WithEmbedderV2(embedder),
			chroma.WithNameSpaceV2("deepseek-rag"),
		)
	default:
		logger.Error("vector db not exist", "VectorDBTypee", *VectorDBType)
		return
	}

	if err != nil {
		logger.Error("get rag store fail", "err", err)
		return
	}

	docs, err := handleKnowledgeBase(ctx, store)
	if err != nil {
		logger.Error("get doc fail", "err", err)
		return
	}

	if len(docs) > 0 {
		_, err = store.AddDocuments(context.Background(), docs)
		if err != nil {
			logger.Error("get save doc fail", "err", err)
			return
		}
	}

}

func handleKnowledgeBase(ctx context.Context, store vectorstores.VectorStore) ([]schema.Document, error) {
	res := make([]schema.Document, 0)

	entries, err := os.ReadDir(*KnowledgePath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".txt") {
			fullPath := filepath.Join(*KnowledgePath, entry.Name())
			f, err := os.Open(fullPath)
			if err != nil {
				logger.Error("read file fail", "err", err)
				continue
			}
			loader := documentloaders.NewText(f)
			splitter := textsplitter.NewRecursiveCharacter(
				textsplitter.WithChunkSize(500),
				textsplitter.WithChunkOverlap(50),
			)

			docs, err := loader.LoadAndSplit(ctx, splitter)
			if err != nil {
				logger.Error("get rag docs fail: %v", err)
				continue
			}

			for _, doc := range docs {
				existingDocs, err := store.SimilaritySearch(context.Background(), doc.PageContent, 1)
				if err == nil && len(existingDocs) > 0 && existingDocs[0].PageContent == doc.PageContent {
					continue
				}

				res = append(res, doc)
			}
		}
	}

	return res, nil

}

func initOpenAIEmbedding() (embeddings.Embedder, error) {
	llm, err := openai.New(
		openai.WithToken(*OpenAIToken),
	)

	if err != nil {
		return nil, err
	}
	embedder, err := embeddings.NewEmbedder(llm)
	if err != nil {
		return nil, err
	}

	return embedder, err
}

func initErnieEmbedding() (embeddings.Embedder, error) {
	llm, err := ernie.New(
		ernie.WithModelName(ernie.ModelNameERNIEBot),
		ernie.WithAKSK(*ErnieAK, *ErnieSK),
	)

	if err != nil {
		return nil, err
	}
	embedder, err := embeddings.NewEmbedder(llm)
	if err != nil {
		return nil, err
	}

	return embedder, err
}

func initGeminiEmbedding(ctx context.Context) (embeddings.Embedder, error) {
	llm, err := googleai.New(ctx,
		googleai.WithAPIKey(*GeminiToken),
	)

	if err != nil {
		return nil, err
	}
	embedder, err := embeddings.NewEmbedder(llm)
	if err != nil {
		return nil, err
	}

	return embedder, err
}
