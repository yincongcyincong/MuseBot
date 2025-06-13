package conf

import (
	"flag"
	"os"
	"strconv"

	"github.com/yincongcyincong/langchaingo/embeddings"
	"github.com/yincongcyincong/langchaingo/vectorstores"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
)

var (
	EmbeddingType *string
	KnowledgePath *string
	VectorDBType  *string

	ChromaURL      *string
	MilvusURL      *string
	WeaviateURL    *string
	WeaviateScheme *string

	Space *string

	ChunkSize    *int
	ChunkOverlap *int

	Store    vectorstores.VectorStore
	Embedder embeddings.Embedder

	DefaultSpliter = []string{"\n\n", "\n", " ", ""}
)

func InitRagConf() {
	EmbeddingType = flag.String("embedding_type", "", "embedding split api: openai gemini ernie")
	KnowledgePath = flag.String("knowledge_path", "./data/knowledge", "knowledge")
	VectorDBType = flag.String("vector_db_type", "chroma", "vector db type: chroma weaviate milvus")

	ChromaURL = flag.String("chroma_url", "http://localhost:8000", "chroma url")
	MilvusURL = flag.String("milvus_url", "http://localhost:19530", "milvus url")
	WeaviateURL = flag.String("weaviate_url", "localhost:8000", "weaviate url localhost:8000")
	WeaviateScheme = flag.String("weaviate_scheme", "http", "weaviate scheme: http")
	Space = flag.String("space", "telegramDeepseekBot", "chroma space")

	ChunkSize = flag.Int("chunk_size", 500, "rag file chunk size")
	ChunkOverlap = flag.Int("chunk_overlap", 50, "rag file chunk overlap")

}

func EnvRagConf() {
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

	if os.Getenv("MILVUS_URL") != "" {
		*MilvusURL = os.Getenv("MILVUS_URL")
	}

	if os.Getenv("WEAVIATE_SCHEME") != "" {
		*WeaviateScheme = os.Getenv("WEAVIATE_SCHEME")
	}

	if os.Getenv("WEAVIATE_URL") != "" {
		*WeaviateURL = os.Getenv("WEAVIATE_URL")
	}

	if os.Getenv("SPACE") != "" {
		*Space = os.Getenv("SPACE")
	}

	if os.Getenv("CHUNK_SIZE") != "" {
		*ChunkSize, _ = strconv.Atoi(os.Getenv("CHUNK_SIZE"))
	}

	if os.Getenv("CHUNK_OVERLAP") != "" {
		*ChunkOverlap, _ = strconv.Atoi(os.Getenv("CHUNK_OVERLAP"))
	}

	logger.Info("RAG_CONF", "EmbeddingType", *EmbeddingType)
	logger.Info("RAG_CONF", "KnowledgePath", *KnowledgePath)
	logger.Info("RAG_CONF", "VectorDBType", *VectorDBType)
	logger.Info("RAG_CONF", "ChromaURL", *ChromaURL)
	logger.Info("RAG_CONF", "ChromaSpace", *Space)
	logger.Info("RAG_CONF", "MilvusURL", *MilvusURL)
	logger.Info("RAG_CONF", "WeaviateURL", *WeaviateURL)
	logger.Info("RAG_CONF", "WeaviateScheme", *WeaviateScheme)
}
