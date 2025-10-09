package conf

import (
	"flag"
	"os"
	"strconv"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/langchaingo/embeddings"
	"github.com/yincongcyincong/langchaingo/vectorstores"
)

type RagConf struct {
	EmbeddingType *string `json:"embedding_type"`
	KnowledgePath *string `json:"knowledge_path"`
	VectorDBType  *string `json:"vector_db_type"`

	ChromaURL      *string `json:"chroma_url"`
	MilvusURL      *string `json:"milvus_url"`
	WeaviateURL    *string `json:"weaviate_url"`
	WeaviateScheme *string `json:"weaviate_scheme"`

	Space *string `json:"space"`

	ChunkSize    *int `json:"chunk_size"`
	ChunkOverlap *int `json:"chunk_overlap"`

	Store          vectorstores.VectorStore `json:"-"`
	Embedder       embeddings.Embedder      `json:"-"`
	MilvusClient   client.Client            `json:"-"`
	WeaviateClient *weaviate.Client         `json:"-"`
}

var (
	RagConfInfo = new(RagConf)

	DefaultSpliter = []string{"\n\n", "\n", " ", ""}
)

func InitRagConf() {
	RagConfInfo.EmbeddingType = flag.String("embedding_type", "", "embedding split api: openai gemini ernie")
	RagConfInfo.KnowledgePath = flag.String("knowledge_path", GetAbsPath("data/knowledge"), "knowledge")
	RagConfInfo.VectorDBType = flag.String("vector_db_type", "milvus", "vector db type: chroma weaviate milvus")

	RagConfInfo.ChromaURL = flag.String("chroma_url", "http://localhost:8000", "chroma url")
	RagConfInfo.MilvusURL = flag.String("milvus_url", "http://localhost:19530", "milvus url")
	RagConfInfo.WeaviateURL = flag.String("weaviate_url", "localhost:8000", "weaviate url localhost:8000")
	RagConfInfo.WeaviateScheme = flag.String("weaviate_scheme", "http", "weaviate scheme: http")
	RagConfInfo.Space = flag.String("space", "MuseBot", "chroma space")

	RagConfInfo.ChunkSize = flag.Int("chunk_size", 500, "rag file chunk size")
	RagConfInfo.ChunkOverlap = flag.Int("chunk_overlap", 50, "rag file chunk overlap")

}

func EnvRagConf() {
	if os.Getenv("EMBEDDING_TYPE") != "" {
		*RagConfInfo.EmbeddingType = os.Getenv("EMBEDDING_TYPE")
	}

	if os.Getenv("KNOWLEDGE_PATH") != "" {
		*RagConfInfo.KnowledgePath = os.Getenv("KNOWLEDGE_PATH")
	}

	if os.Getenv("VECTOR_DB_TYPE") != "" {
		*RagConfInfo.VectorDBType = os.Getenv("VECTOR_DB_TYPE")
	}

	if os.Getenv("CHROMA_URL") != "" {
		*RagConfInfo.ChromaURL = os.Getenv("CHROMA_URL")
	}

	if os.Getenv("MILVUS_URL") != "" {
		*RagConfInfo.MilvusURL = os.Getenv("MILVUS_URL")
	}

	if os.Getenv("WEAVIATE_SCHEME") != "" {
		*RagConfInfo.WeaviateScheme = os.Getenv("WEAVIATE_SCHEME")
	}

	if os.Getenv("WEAVIATE_URL") != "" {
		*RagConfInfo.WeaviateURL = os.Getenv("WEAVIATE_URL")
	}

	if os.Getenv("SPACE") != "" {
		*RagConfInfo.Space = os.Getenv("SPACE")
	}

	if os.Getenv("CHUNK_SIZE") != "" {
		*RagConfInfo.ChunkSize, _ = strconv.Atoi(os.Getenv("CHUNK_SIZE"))
	}

	if os.Getenv("CHUNK_OVERLAP") != "" {
		*RagConfInfo.ChunkOverlap, _ = strconv.Atoi(os.Getenv("CHUNK_OVERLAP"))
	}

	logger.Info("RAG_CONF", "EmbeddingType", *RagConfInfo.EmbeddingType)
	logger.Info("RAG_CONF", "KnowledgePath", *RagConfInfo.KnowledgePath)
	logger.Info("RAG_CONF", "VectorDBType", *RagConfInfo.VectorDBType)
	logger.Info("RAG_CONF", "ChromaURL", *RagConfInfo.ChromaURL)
	logger.Info("RAG_CONF", "ChromaSpace", *RagConfInfo.Space)
	logger.Info("RAG_CONF", "MilvusURL", *RagConfInfo.MilvusURL)
	logger.Info("RAG_CONF", "WeaviateURL", *RagConfInfo.WeaviateURL)
	logger.Info("RAG_CONF", "WeaviateScheme", *RagConfInfo.WeaviateScheme)
}
