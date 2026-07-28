package conf

import (
	"flag"
	"os"
	"strconv"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/yincongcyincong/langchaingo/embeddings"
	"github.com/yincongcyincong/langchaingo/vectorstores"
)

type RagConf struct {
	EmbeddingType string `json:"embedding_type"`
	KnowledgePath string `json:"knowledge_path"`
	VectorDBType  string `json:"vector_db_type"`

	ChromaURL      string `json:"chroma_url"`
	MilvusURL      string `json:"milvus_url"`
	WeaviateURL    string `json:"weaviate_url"`
	WeaviateScheme string `json:"weaviate_scheme"`
	QdrantHost     string `json:"qdrant_host"`
	QdrantPort     int    `json:"qdrant_port"`
	QdrantAPIKey   string `json:"qdrant_api_key"`
	QdrantUseTLS   bool   `json:"qdrant_use_tls"`

	Space string `json:"space"`

	ChunkSize    int `json:"chunk_size"`
	ChunkOverlap int `json:"chunk_overlap"`

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
	flag.StringVar(&RagConfInfo.EmbeddingType, "embedding_type", "", "embedding split api: openai gemini ernie")
	flag.StringVar(&RagConfInfo.KnowledgePath, "knowledge_path", GetAbsPath("data/knowledge"), "knowledge")
	flag.StringVar(&RagConfInfo.VectorDBType, "vector_db_type", "milvus", "vector db type: weaviate milvus qdrant")

	flag.StringVar(&RagConfInfo.ChromaURL, "chroma_url", "http://localhost:8000", "chroma url")
	flag.StringVar(&RagConfInfo.MilvusURL, "milvus_url", "http://localhost:19530", "milvus url")
	flag.StringVar(&RagConfInfo.WeaviateURL, "weaviate_url", "localhost:8000", "weaviate url localhost:8000")
	flag.StringVar(&RagConfInfo.WeaviateScheme, "weaviate_scheme", "http", "weaviate scheme: http")
	flag.StringVar(&RagConfInfo.QdrantHost, "qdrant_host", "localhost", "qdrant gRPC host")
	flag.IntVar(&RagConfInfo.QdrantPort, "qdrant_port", 6334, "qdrant gRPC port")
	flag.StringVar(&RagConfInfo.QdrantAPIKey, "qdrant_api_key", "", "qdrant api key")
	flag.BoolVar(&RagConfInfo.QdrantUseTLS, "qdrant_use_tls", false, "use TLS for qdrant gRPC")
	flag.StringVar(&RagConfInfo.Space, "space", "MuseBot", "vector db collection or space")

	flag.IntVar(&RagConfInfo.ChunkSize, "chunk_size", 500, "rag file chunk size")
	flag.IntVar(&RagConfInfo.ChunkOverlap, "chunk_overlap", 50, "rag file chunk overlap")

}

func EnvRagConf() {
	if os.Getenv("EMBEDDING_TYPE") != "" {
		RagConfInfo.EmbeddingType = os.Getenv("EMBEDDING_TYPE")
	}

	if os.Getenv("KNOWLEDGE_PATH") != "" {
		RagConfInfo.KnowledgePath = os.Getenv("KNOWLEDGE_PATH")
	}

	if os.Getenv("VECTOR_DB_TYPE") != "" {
		RagConfInfo.VectorDBType = os.Getenv("VECTOR_DB_TYPE")
	}

	if os.Getenv("CHROMA_URL") != "" {
		RagConfInfo.ChromaURL = os.Getenv("CHROMA_URL")
	}

	if os.Getenv("MILVUS_URL") != "" {
		RagConfInfo.MilvusURL = os.Getenv("MILVUS_URL")
	}

	if os.Getenv("WEAVIATE_SCHEME") != "" {
		RagConfInfo.WeaviateScheme = os.Getenv("WEAVIATE_SCHEME")
	}

	if os.Getenv("WEAVIATE_URL") != "" {
		RagConfInfo.WeaviateURL = os.Getenv("WEAVIATE_URL")
	}

	if os.Getenv("QDRANT_HOST") != "" {
		RagConfInfo.QdrantHost = os.Getenv("QDRANT_HOST")
	}

	if os.Getenv("QDRANT_PORT") != "" {
		RagConfInfo.QdrantPort, _ = strconv.Atoi(os.Getenv("QDRANT_PORT"))
	}

	if os.Getenv("QDRANT_API_KEY") != "" {
		RagConfInfo.QdrantAPIKey = os.Getenv("QDRANT_API_KEY")
	}

	if os.Getenv("QDRANT_USE_TLS") != "" {
		RagConfInfo.QdrantUseTLS, _ = strconv.ParseBool(os.Getenv("QDRANT_USE_TLS"))
	}

	if os.Getenv("SPACE") != "" {
		RagConfInfo.Space = os.Getenv("SPACE")
	}

	if os.Getenv("CHUNK_SIZE") != "" {
		RagConfInfo.ChunkSize, _ = strconv.Atoi(os.Getenv("CHUNK_SIZE"))
	}

	if os.Getenv("CHUNK_OVERLAP") != "" {
		RagConfInfo.ChunkOverlap, _ = strconv.Atoi(os.Getenv("CHUNK_OVERLAP"))
	}
}
