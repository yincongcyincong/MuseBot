package rag

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"
	
	"github.com/cohesion-org/deepseek-go"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/yincongcyincong/langchaingo/documentloaders"
	"github.com/yincongcyincong/langchaingo/embeddings"
	"github.com/yincongcyincong/langchaingo/llms"
	"github.com/yincongcyincong/langchaingo/llms/ernie"
	"github.com/yincongcyincong/langchaingo/llms/googleai"
	"github.com/yincongcyincong/langchaingo/llms/openai"
	"github.com/yincongcyincong/langchaingo/schema"
	"github.com/yincongcyincong/langchaingo/textsplitter"
	"github.com/yincongcyincong/langchaingo/vectorstores/chroma"
	"github.com/yincongcyincong/langchaingo/vectorstores/milvus"
	"github.com/yincongcyincong/langchaingo/vectorstores/weaviate"
	"github.com/yincongcyincong/telegram-deepseek-bot/conf"
	"github.com/yincongcyincong/telegram-deepseek-bot/db"
	"github.com/yincongcyincong/telegram-deepseek-bot/llm"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
	"github.com/yincongcyincong/telegram-deepseek-bot/utils"
)

type Rag struct {
	Client *deepseek.Client
	
	LLM *llm.LLM
}

func NewRag(options ...llm.Option) *Rag {
	dp := &Rag{
		Client: deepseek.NewClient(*conf.BaseConfInfo.DeepseekToken),
		
		LLM: llm.NewLLM(options...),
	}
	
	for _, o := range options {
		o(dp.LLM)
	}
	return dp
}

func (l *Rag) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return llms.GenerateFromSinglePrompt(ctx, l, prompt, options...)
}

func (l *Rag) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	opts := &llms.CallOptions{}
	for _, opt := range options {
		opt(opts)
	}
	
	doc, err := conf.RagConfInfo.Store.SimilaritySearch(ctx, l.LLM.Content, 3)
	if err != nil {
		logger.Error("request vector db fail", "err", err)
	}
	if len(doc) != 0 {
		tmpContent := ""
		for _, msg := range messages {
			for _, part := range msg.Parts {
				tmpContent += part.(llms.TextContent).Text
			}
		}
		l.LLM.Content = tmpContent
	}
	
	err = l.LLM.CallLLM()
	if err != nil {
		logger.Error("error calling DeepSeek API", "err", err)
		return nil, errors.New("error calling DeepSeek API")
	}
	
	resp := &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{
				Content: l.LLM.WholeContent,
			},
		},
	}
	
	return resp, nil
}

func InitRag() {
	if *conf.RagConfInfo.EmbeddingType == "" || *conf.RagConfInfo.VectorDBType == "" {
		return
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	
	var err error
	switch *conf.RagConfInfo.EmbeddingType {
	case "openai":
		conf.RagConfInfo.Embedder, err = initOpenAIEmbedding()
	case "gemini":
		conf.RagConfInfo.Embedder, err = initGeminiEmbedding(ctx)
	case "ernie":
		conf.RagConfInfo.Embedder, err = initErnieEmbedding()
	default:
		logger.Error("embedding type not exist", "embedding type", *conf.RagConfInfo.EmbeddingType)
		return
	}
	
	if err != nil {
		logger.Error("init embedding fail", "err", err)
		return
	}
	
	switch *conf.RagConfInfo.VectorDBType {
	case "chroma":
		conf.RagConfInfo.Store, err = chroma.NewV2(
			chroma.WithChromaURLV2(*conf.RagConfInfo.ChromaURL),
			chroma.WithEmbedderV2(conf.RagConfInfo.Embedder),
			chroma.WithNameSpaceV2(*conf.RagConfInfo.Space),
		)
	case "milvus":
		idx, err := entity.NewIndexAUTOINDEX(entity.L2)
		if err != nil {
			logger.Error("get index fail", "err", err)
			return
		}
		
		conf.RagConfInfo.Store, err = milvus.New(ctx, client.Config{
			Address: *conf.RagConfInfo.MilvusURL,
		}, milvus.WithCollectionName(*conf.RagConfInfo.Space),
			milvus.WithEmbedder(conf.RagConfInfo.Embedder),
			milvus.WithIndex(idx))
	case "weaviate":
		conf.RagConfInfo.Store, err = weaviate.New(
			weaviate.WithEmbedder(conf.RagConfInfo.Embedder),
			weaviate.WithScheme(*conf.RagConfInfo.WeaviateScheme),
			weaviate.WithHost(*conf.RagConfInfo.WeaviateURL),
			weaviate.WithIndexName("Text"))
	default:
		logger.Error("vector db not exist", "VectorDBTypee", *conf.RagConfInfo.VectorDBType)
		return
	}
	
	if err != nil {
		logger.Error("get rag store fail", "err", err)
		return
	}
	
	docs, err := handleKnowledgeBase(ctx)
	if err != nil {
		logger.Error("get doc fail", "err", err)
		return
	}
	
	if len(docs) > 0 {
		_, err = conf.RagConfInfo.Store.AddDocuments(context.Background(), docs)
		if err != nil {
			logger.Error("get save doc fail", "err", err)
			return
		}
	}
	
}

func handleKnowledgeBase(ctx context.Context) ([]schema.Document, error) {
	res := make([]schema.Document, 0)
	
	entries, err := os.ReadDir(*conf.RagConfInfo.KnowledgePath)
	if err != nil {
		return nil, err
	}
	
	for _, entry := range entries {
		if !entry.IsDir() {
			var docs []schema.Document
			switch {
			case strings.HasSuffix(strings.ToLower(entry.Name()), ".txt"):
				docs, err = handleTextDoc(ctx, entry)
				if err != nil {
					logger.Error("handle text doc fail", "err", err)
				}
			case strings.HasSuffix(strings.ToLower(entry.Name()), ".pdf"):
				docs, err = handlePDFDoc(ctx, entry)
				if err != nil {
					logger.Error("handle pdf doc fail", "err", err)
				}
			case strings.HasSuffix(strings.ToLower(entry.Name()), ".csv"):
				docs, err = handleCSVDoc(ctx, entry)
				if err != nil {
					logger.Error("handle csv doc fail", "err", err)
				}
			case strings.HasSuffix(strings.ToLower(entry.Name()), ".html"):
				docs, err = handleHTMLDoc(ctx, entry)
				if err != nil {
					logger.Error("handle html doc fail", "err", err)
				}
			}
			if len(docs) > 0 {
				res = append(res, docs...)
			}
		}
	}
	
	return res, nil
	
}

func initOpenAIEmbedding() (embeddings.Embedder, error) {
	llm, err := openai.New(
		openai.WithToken(*conf.BaseConfInfo.OpenAIToken),
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
		ernie.WithAKSK(*conf.BaseConfInfo.ErnieAK, *conf.BaseConfInfo.ErnieSK),
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
		googleai.WithAPIKey(*conf.BaseConfInfo.GeminiToken),
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

func getFileResource(entry os.DirEntry) (*os.File, error) {
	fullPath := filepath.Join(*conf.RagConfInfo.KnowledgePath, entry.Name())
	
	fileMd5, err := utils.FileToMd5(fullPath)
	if err != nil {
		logger.Error("file to md5 fail", "err", err)
		return nil, err
	}
	
	fileInfos, err := db.GetRagFileByFileMd5(fileMd5)
	if err != nil {
		logger.Error("get file from db fail", "err", err)
		return nil, err
	}
	
	if len(fileInfos) > 0 {
		logger.Info("file exist", "path", fullPath)
		return nil, nil
	}
	
	_, err = db.InsertRagFile(entry.Name(), fileMd5)
	if err != nil {
		logger.Error("insert rag file fail", "err", err)
	}
	
	return os.Open(fullPath)
}

func handleTextDoc(ctx context.Context, entry os.DirEntry) ([]schema.Document, error) {
	f, err := getFileResource(entry)
	if err != nil {
		logger.Error("read file fail", "err", err)
		return nil, err
	}
	if f == nil {
		return nil, nil
	}
	defer f.Close()
	
	loader := documentloaders.NewText(f)
	return saveDocIntoStore(ctx, loader)
}

func handlePDFDoc(ctx context.Context, entry os.DirEntry) ([]schema.Document, error) {
	f, err := getFileResource(entry)
	if err != nil {
		logger.Error("read file fail", "err", err)
		return nil, err
	}
	if f == nil {
		return nil, nil
	}
	defer f.Close()
	
	finfo, err := f.Stat()
	if err != nil {
		logger.Error("get file stat fail", "err", err)
		return nil, err
	}
	loader := documentloaders.NewPDF(f, finfo.Size())
	return saveDocIntoStore(ctx, loader)
}

func handleCSVDoc(ctx context.Context, entry os.DirEntry) ([]schema.Document, error) {
	f, err := getFileResource(entry)
	if err != nil {
		logger.Error("read file fail", "err", err)
		return nil, err
	}
	if f == nil {
		return nil, nil
	}
	defer f.Close()
	
	loader := documentloaders.NewCSV(f)
	return saveDocIntoStore(ctx, loader)
}

func handleHTMLDoc(ctx context.Context, entry os.DirEntry) ([]schema.Document, error) {
	f, err := getFileResource(entry)
	if err != nil {
		logger.Error("read file fail", "err", err)
		return nil, err
	}
	if f == nil {
		return nil, nil
	}
	defer f.Close()
	
	loader := documentloaders.NewHTML(f)
	return saveDocIntoStore(ctx, loader)
}

func saveDocIntoStore(ctx context.Context, loader documentloaders.Loader) ([]schema.Document, error) {
	splitter := textsplitter.NewRecursiveCharacter(
		textsplitter.WithChunkSize(*conf.RagConfInfo.ChunkSize),
		textsplitter.WithChunkOverlap(*conf.RagConfInfo.ChunkOverlap),
		textsplitter.WithSeparators(conf.DefaultSpliter),
	)
	
	docs, err := loader.LoadAndSplit(ctx, splitter)
	if err != nil {
		logger.Error("get rag docs fail: %v", err)
		return nil, err
	}
	
	return docs, nil
}
