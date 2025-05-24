package mcpagent

import "fmt"

// 命令行通知器
type CliNotifier struct {
}

func (n *CliNotifier) OnMessage(msg string) {
	fmt.Println(msg)
}

func (n *CliNotifier) OnResult(msg string) {
	fmt.Println(msg)
}

func (n *CliNotifier) OnError(err error) {
	fmt.Println(err)
}
