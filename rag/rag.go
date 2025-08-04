package rag

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	db_weaviate "github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/yincongcyincong/langchaingo/documentloaders"
	"github.com/yincongcyincong/langchaingo/embeddings"
	"github.com/yincongcyincong/langchaingo/llms"
	"github.com/yincongcyincong/langchaingo/llms/ernie"
	"github.com/yincongcyincong/langchaingo/llms/googleai"
	"github.com/yincongcyincong/langchaingo/llms/openai"
	"github.com/yincongcyincong/langchaingo/schema"
	"github.com/yincongcyincong/langchaingo/textsplitter"
	"github.com/yincongcyincong/langchaingo/vectorstores/milvus"
	"github.com/yincongcyincong/langchaingo/vectorstores/weaviate"
	"github.com/yincongcyincong/telegram-deepseek-bot/conf"
	"github.com/yincongcyincong/telegram-deepseek-bot/db"
	"github.com/yincongcyincong/telegram-deepseek-bot/llm"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
	"github.com/yincongcyincong/telegram-deepseek-bot/utils"
	"gopkg.in/fsnotify.v1"
)

const (
	weaviateIndexName = "TelegramDeepseekBot"
)

type Rag struct {
	LLM *llm.LLM
}

func NewRag(options ...llm.Option) *Rag {
	dp := &Rag{
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
		llm.WithContent(tmpContent)(l.LLM)
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
	//case "chroma":
	//	conf.RagConfInfo.Store, err = chroma.NewV2(
	//		chroma.WithChromaURLV2(*conf.RagConfInfo.ChromaURL),
	//		chroma.WithEmbedderV2(conf.RagConfInfo.Embedder),
	//		chroma.WithNameSpaceV2(*conf.RagConfInfo.Space),
	//	)
	case "milvus":
		idx, err := entity.NewIndexAUTOINDEX(entity.L2)
		if err != nil {
			logger.Error("get index fail", "err", err)
			return
		}
		clientConf := client.Config{
			Address: *conf.RagConfInfo.MilvusURL,
		}
		conf.RagConfInfo.Store, err = milvus.New(ctx, clientConf,
			milvus.WithCollectionName(*conf.RagConfInfo.Space),
			milvus.WithEmbedder(conf.RagConfInfo.Embedder),
			milvus.WithIndex(idx),
			milvus.WithDropOld())
		conf.RagConfInfo.MilvusClient, _ = client.NewClient(ctx, clientConf)
	case "weaviate":
		conf.RagConfInfo.Store, err = weaviate.New(
			weaviate.WithEmbedder(conf.RagConfInfo.Embedder),
			weaviate.WithScheme(*conf.RagConfInfo.WeaviateScheme),
			weaviate.WithHost(*conf.RagConfInfo.WeaviateURL),
			weaviate.WithIndexName(weaviateIndexName))
		
		conf.RagConfInfo.WeaviateClient, _ = db_weaviate.NewClient(db_weaviate.Config{
			Scheme: *conf.RagConfInfo.WeaviateScheme,
			Host:   *conf.RagConfInfo.WeaviateURL,
		})
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
		insertVectorDb(ctx, docs)
	}
	
	go CheckDirChange()
	
}

func insertVectorDb(ctx context.Context, docs []schema.Document) {
	ids, err := conf.RagConfInfo.Store.AddDocuments(ctx, docs)
	if err != nil {
		logger.Error("get save doc fail", "err", err)
		return
	}
	
	fileVectorIds := make(map[string]string)
	for i := range docs {
		fileMd5 := docs[i].Metadata["file_md5"].(string)
		fileVectorIds[fileMd5] += ids[i] + ","
	}
	
	for fileMd5, vectorIds := range fileVectorIds {
		err = db.UpdateVectorIdByFileMd5(fileMd5, strings.TrimRight(vectorIds, ","))
		if err != nil {
			logger.Error("update vector id fail", "err", err)
		}
	}
}

func handleKnowledgeBase(ctx context.Context) ([]schema.Document, error) {
	entries, err := os.ReadDir(*conf.RagConfInfo.KnowledgePath)
	if err != nil {
		return nil, err
	}
	
	return handleEntry(ctx, entries)
	
}

func handleEntry(ctx context.Context, entries []os.DirEntry) ([]schema.Document, error) {
	var err error
	res := make([]schema.Document, 0)
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
	llmEmbedder, err := openai.New(
		openai.WithToken(*conf.BaseConfInfo.OpenAIToken),
	)
	
	if err != nil {
		return nil, err
	}
	embedder, err := embeddings.NewEmbedder(llmEmbedder)
	if err != nil {
		return nil, err
	}
	
	return embedder, err
}

func initErnieEmbedding() (embeddings.Embedder, error) {
	llmEmbedder, err := ernie.New(
		ernie.WithModelName(ernie.ModelNameERNIEBot),
		ernie.WithAKSK(*conf.BaseConfInfo.ErnieAK, *conf.BaseConfInfo.ErnieSK),
	)
	
	if err != nil {
		return nil, err
	}
	embedder, err := embeddings.NewEmbedder(llmEmbedder)
	if err != nil {
		return nil, err
	}
	
	return embedder, err
}

func initGeminiEmbedding(ctx context.Context) (embeddings.Embedder, error) {
	llmEmbedder, err := googleai.New(ctx,
		googleai.WithAPIKey(*conf.BaseConfInfo.GeminiToken),
	)
	
	if err != nil {
		return nil, err
	}
	embedder, err := embeddings.NewEmbedder(llmEmbedder)
	if err != nil {
		return nil, err
	}
	
	return embedder, err
}

func getFileResource(entry os.DirEntry) (*os.File, string, error) {
	fullPath := filepath.Join(*conf.RagConfInfo.KnowledgePath, entry.Name())
	
	fileMd5, err := utils.FileToMd5(fullPath)
	if err != nil {
		logger.Error("file to md5 fail", "err", err)
		return nil, "", err
	}
	
	fileInfos, err := db.GetRagFileByFileMd5(fileMd5)
	if err != nil {
		logger.Error("get file from db fail", "err", err)
		return nil, "", err
	}
	
	if len(fileInfos) > 0 {
		logger.Info("file exist", "path", fullPath)
		return nil, "", nil
	}
	
	err = db.DeleteRagFileByFileName(entry.Name())
	if err != nil {
		logger.Error("delete file from db fail", "err", err)
	}
	
	_, err = db.InsertRagFile(entry.Name(), fileMd5)
	if err != nil {
		logger.Error("insert rag file fail", "err", err)
	}
	
	f, err := os.Open(fullPath)
	return f, fileMd5, err
}

func handleTextDoc(ctx context.Context, entry os.DirEntry) ([]schema.Document, error) {
	f, fMd5, err := getFileResource(entry)
	if err != nil {
		logger.Error("read file fail", "err", err)
		return nil, err
	}
	if f == nil {
		return nil, nil
	}
	defer f.Close()
	
	loader := documentloaders.NewText(f)
	return saveDocIntoStore(ctx, loader, fMd5, entry)
}

func handlePDFDoc(ctx context.Context, entry os.DirEntry) ([]schema.Document, error) {
	f, fMd5, err := getFileResource(entry)
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
	return saveDocIntoStore(ctx, loader, fMd5, entry)
}

func handleCSVDoc(ctx context.Context, entry os.DirEntry) ([]schema.Document, error) {
	f, fMd5, err := getFileResource(entry)
	if err != nil {
		logger.Error("read file fail", "err", err)
		return nil, err
	}
	if f == nil {
		return nil, nil
	}
	defer f.Close()
	
	loader := documentloaders.NewCSV(f)
	return saveDocIntoStore(ctx, loader, fMd5, entry)
}

func handleHTMLDoc(ctx context.Context, entry os.DirEntry) ([]schema.Document, error) {
	f, fMd5, err := getFileResource(entry)
	if err != nil {
		logger.Error("read file fail", "err", err)
		return nil, err
	}
	if f == nil {
		return nil, nil
	}
	defer f.Close()
	
	loader := documentloaders.NewHTML(f)
	return saveDocIntoStore(ctx, loader, fMd5, entry)
}

func saveDocIntoStore(ctx context.Context, loader documentloaders.Loader, fMd5 string, entry os.DirEntry) ([]schema.Document, error) {
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
	
	for _, doc := range docs {
		doc.Metadata["file_name"] = entry.Name()
		doc.Metadata["file_md5"] = fMd5
	}
	
	return docs, nil
}

func CheckDirChange() {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("CheckDirChange panic", "err", err)
		}
	}()
	
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Error("create watcher fail", "err", err)
		return
	}
	defer watcher.Close()
	
	// 监控当前目录
	err = watcher.Add(*conf.RagConfInfo.KnowledgePath)
	if err != nil {
		logger.Error("add watcher fail", "err", err)
		return
	}
	
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			insertNewDoc(event)
		case err, ok := <-watcher.Errors:
			if !ok {
				logger.Error("watcher channel closed")
				return
			}
			logger.Error("watcher error", "err", err)
		}
	}
	
}

func insertNewDoc(event fsnotify.Event) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	switch {
	case event.Op&fsnotify.Create == fsnotify.Create:
		logger.Info("rag dir changed", "event", event.Name, "op", "create")
		InsertDoc(ctx, event)
	case event.Op&fsnotify.Write == fsnotify.Write:
		logger.Info("rag dir changed", "event", event.Name, "op", "write")
		DeleteDoc(ctx, event)
		InsertDoc(ctx, event)
	case event.Op&fsnotify.Remove == fsnotify.Remove:
		logger.Info("rag dir changed", "event", event.Name, "op", "remove")
		DeleteDoc(ctx, event)
	}
	
}

func DeleteDoc(ctx context.Context, event fsnotify.Event) {
	fileName := filepath.Base(event.Name)
	fileDbInfo, err := db.GetRagFileByFileName(fileName)
	if err != nil {
		logger.Error("get file db info fail", "err", err)
		return
	}
	if fileDbInfo != nil && len(fileDbInfo) > 0 {
		err = DeleteStoreData(ctx, fileDbInfo[0].VectorId)
		if err != nil {
			logger.Error("delete doc fail", "err", err)
			return
		}
		err = db.DeleteRagFileByFileName(fileDbInfo[0].FileName)
		if err != nil {
			logger.Error("delete doc in db fail", "err", err)
			return
		}
	}
}

func InsertDoc(ctx context.Context, event fsnotify.Event) {
	fileInfo, err := os.Stat(event.Name)
	if err != nil {
		logger.Error("stat file fail", "err", err)
		return
	}
	entry, err := findDirEntry(*conf.RagConfInfo.KnowledgePath, fileInfo.Name())
	if err != nil {
		logger.Error("find dir entry fail", "err", err)
		return
	}
	docs, err := handleEntry(ctx, []os.DirEntry{entry})
	if err != nil {
		logger.Error("handle entry fail", "err", err)
		return
	}
	if len(docs) > 0 {
		insertVectorDb(ctx, docs)
	}
}

func findDirEntry(path string, filename string) (os.DirEntry, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	
	for _, entry := range entries {
		if entry.Name() == filename {
			return entry, nil
		}
	}
	return nil, fmt.Errorf("file not exist: %s", filename)
}

func DeleteStoreData(ctx context.Context, vectorIds string) error {
	var err error
	switch *conf.RagConfInfo.VectorDBType {
	case "weaviate":
		for _, vectorId := range strings.Split(vectorIds, ",") {
			err = conf.RagConfInfo.WeaviateClient.Data().Deleter().
				WithClassName(weaviateIndexName).
				WithID(vectorId).
				Do(ctx)
			if err != nil {
				logger.Error("delete store data fail", "err", err)
			}
		}
	
	case "milvus":
		for _, vectorId := range strings.Split(vectorIds, ",") {
			expr := fmt.Sprintf(`pk == %s`, vectorId)
			err = conf.RagConfInfo.MilvusClient.Delete(ctx, *conf.RagConfInfo.Space, "", expr)
		}
	}
	
	return nil
}
