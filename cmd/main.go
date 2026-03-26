package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"ai90/internal/config"
	"ai90/internal/service"
)

// deprecated 使用webserver
func main() {
	// 加载配置
	config.GlobalCfg = *config.Load()

	// 解析命令行参数
	var singleMode bool
	var question string
	flag.BoolVar(&singleMode, "s", false, "单次提问模式（非交互式）")
	flag.StringVar(&question, "q", "", "要提问的问题（与 -s 一起使用）")
	flag.Parse()

	// 打印配置信息
	fmt.Println("=== K8s AI 助手 (项目: ai90) ===")
	fmt.Printf("配置: Ollama=%s, Chroma=%s, Collection=%s\n",
		config.GlobalCfg.Ollama.URL, config.GlobalCfg.Chroma.URL, config.GlobalCfg.Chroma.CollectionName)
	fmt.Printf("模型: LLM=%s, Embedding=%s\n", config.GlobalCfg.Model.LLM, config.GlobalCfg.Model.Embedding)
	fmt.Println()

	// 单次提问模式
	if singleMode {
		if question == "" {
			// 从剩余参数获取问题
			args := flag.Args()
			if len(args) > 0 {
				question = strings.Join(args, " ")
			}
		}

		if question == "" {
			fmt.Println("错误: 单次模式需要提供问题，使用 -q \"问题内容\" 或直接在命令后输入问题")
			os.Exit(1)
		}

		executeQuery(question)
		return
	}

	// 交互式模式
	runInteractiveMode()
}

// runInteractiveMode 运行交互式模式
func runInteractiveMode() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("进入交互模式，输入 'exit' 或 'quit' 退出")
	fmt.Println("输入 'help' 查看帮助")
	fmt.Println()

	for {
		fmt.Print("📝 请输入您的问题: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("读取输入错误: %v\n", err)
			continue
		}

		input = strings.TrimSpace(input)

		// 处理特殊命令
		switch strings.ToLower(input) {
		case "exit", "quit", "q":
			fmt.Println("👋 再见！")
			return
		case "help", "h":
			printHelp()
			continue
		case "":
			continue
		}

		// 执行查询
		executeQuery(input)
	}
}

// executeQuery 执行RAG查询并输出结果
func executeQuery(question string) {
	fmt.Println("\n🤖 正在思考...")
	fmt.Println(strings.Repeat("-", 50))

	// 执行 RAG 流程
	answer, err := service.ExecuteRAG(question)
	if err != nil {
		fmt.Printf("❌ 错误: %v\n", err)
		return
	}

	fmt.Println("💡 AI 回答:")
	fmt.Println(answer)
	fmt.Println(strings.Repeat("-", 50))
	fmt.Println()
}

// printHelp 打印帮助信息
func printHelp() {
	fmt.Println("\n📖 帮助信息:")
	fmt.Println("  - 直接输入问题获取AI回答")
	fmt.Println("  - exit/quit/q: 退出程序")
	fmt.Println("  - help/h: 显示此帮助")
	fmt.Println()
	fmt.Println("命令行参数:")
	fmt.Println("  -s: 单次提问模式（非交互式）")
	fmt.Println("  -q \"问题\": 指定问题内容")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  ai90                    # 进入交互模式")
	fmt.Println("  ai90 -s -q \"什么是Pod?\"  # 单次提问")
	fmt.Println("  ai90 -s 什么是Pod?       # 单次提问（简写）")
	fmt.Println()
}
