package rag

import (
	"context"
	"fmt"
	"github.com/cohesion-org/deepseek-go"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ernie"
	"github.com/tmc/langchaingo/vectorstores"
	"log"
	"os"
	"strings"

	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/textsplitter"
	"github.com/tmc/langchaingo/vectorstores/chroma"
)

func execute() {

	// 2. 初始化嵌入模型 (这里使用OpenAI，你也可以用其他)
	//embedderLLM, err := openai.New(
	//	openai.WithToken(os.Getenv("OPENAI_API_KEY")),
	//)

	llm, err := ernie.New(
		ernie.WithModelName(ernie.ModelNameERNIEBot),
		ernie.WithAKSK("", ""),
	)
	if err != nil {
		log.Fatal(err)
	}
	embedder, err := embeddings.NewEmbedder(llm)
	if err != nil {
		log.Fatal(err)
	}

	// 3. 加载和分割文档
	text := "DeepSeek是一家专注于人工智能技术的公司..."
	loader := documentloaders.NewText(strings.NewReader(text))
	splitter := textsplitter.NewRecursiveCharacter(
		textsplitter.WithChunkSize(500),
		textsplitter.WithChunkOverlap(50),
	)
	docs, err := loader.LoadAndSplit(context.Background(), splitter)
	if err != nil {
		log.Fatal(err)
	}

	// 4. 初始化向量存储
	store, err := chroma.New(
		chroma.WithChromaURL("http://localhost:8000"),
		chroma.WithEmbedder(embedder),
		chroma.WithNameSpace("deepseek-rag"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// 5. 添加文档到向量存储
	_, err = store.AddDocuments(context.Background(), docs)
	if err != nil {
		log.Fatal(err)
	}

	// --------------------

	// 1. 初始化DeepSeek LLM
	deepseekAPIKey := os.Getenv("DEEPSEEK_API_KEY")
	dsLLM := NewDeepSeekLLM(deepseekAPIKey)

	// 7. 创建RAG链
	qaChain := chains.NewRetrievalQAFromLLM(
		dsLLM,
		vectorstores.ToRetriever(store, 3),
	)

	// 8. 执行查询
	question := "DeepSeek公司的主要业务是什么？"
	answer, err := chains.Run(context.Background(), qaChain, question)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("问题: %s\n答案: %s\n", question, answer)
}

type DeepSeekLLM struct {
	Client *deepseek.Client
	Model  string
}

func NewDeepSeekLLM(apiKey string) *DeepSeekLLM {
	return &DeepSeekLLM{
		Client: deepseek.NewClient(apiKey),
		Model:  "deepseek-chat",
	}
}

func (l *DeepSeekLLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return llms.GenerateFromSinglePrompt(ctx, l, prompt, options...)
}

func (l *DeepSeekLLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {

	opts := &llms.CallOptions{}
	for _, opt := range options {
		opt(opts)
	}

	// Assume we get a single text message
	msg0 := messages[0]
	part := msg0.Parts[0]
	result, err := l.Client.CreateChatCompletion(ctx, &deepseek.ChatCompletionRequest{
		Messages:    []deepseek.ChatCompletionMessage{{Role: "user", Content: part.(llms.TextContent).Text}},
		Temperature: float32(opts.Temperature),
		TopP:        float32(opts.TopP),
	})
	if err != nil {
		return nil, err
	}
	if len(result.Choices) == 0 {
		err = fmt.Errorf("%w, error_code:%v, erro_msg:%v, id:%v", result)
		return nil, err
	}

	resp := &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{
				Content: result.Choices[0].Message.Content,
			},
		},
	}

	return resp, nil
}
