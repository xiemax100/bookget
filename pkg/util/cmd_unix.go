//go:build !windows

package util

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"time"
)

const (
	TH32CS_SNAPPROCESS = 0x00000002
	MAX_PATH           = 260
)

type ProcessEntry32 struct {
	Size            uint32
	CntUsage        uint32
	ProcessID       uint32
	DefaultHeapID   uintptr
	ModuleID        uint32
	CntThreads      uint32
	ParentProcessID uint32
	PriClassBase    int32
	Flags           uint32
	ExeFile         [MAX_PATH]uint16 // 使用 uint16 而不是 byte
}

func RunCommand(ctx context.Context, text string) error {
	fmt.Println(text)
	var cmd *exec.Cmd
	if os.PathSeparator == '\\' {
		cmd = exec.CommandContext(ctx, "cmd", "/c", text)
	} else {
		cmd = exec.CommandContext(ctx, "bash", "-c", text)
	}
	//捕获标准输出
	stdout, err := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout
	if err != nil {
		return err
	}
	// 执行命令cmd.CombinedOutput(),且捕获输出
	//output, err = cmd.CombinedOutput()
	if err = cmd.Start(); err != nil {
		return err
	}
	readout := bufio.NewReader(stdout)
	GetOutput(readout)
	if err = cmd.Wait(); err != nil {
		return err
	}
	return nil
}

func GetOutput(reader *bufio.Reader) {
	var sumOutput string //统计屏幕的全部输出内容
	outputBytes := make([]byte, 200)
	for {
		n, err := reader.Read(outputBytes) //获取屏幕的实时输出(并不是按照回车分割，所以要结合sumOutput)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println(err)
			sumOutput += err.Error()
		}
		output := string(outputBytes[:n])
		fmt.Print(output) //输出屏幕内容
		sumOutput += output
	}
	return
}

// PrintSleepTime 打印0-60秒等待
func PrintSleepTime(sec int) {
	if sec <= 0 || sec > 60 {
		return
	}
	fmt.Println()
	for t := sec; t > 0; t-- {
		seconds := strconv.Itoa(int(t))
		if t < 10 {
			seconds = fmt.Sprintf("0%d", t)
		}
		fmt.Printf("\rplease wait.... [00:%s of appr. Max %d sec]", seconds, sec)
		time.Sleep(time.Second)
	}
	fmt.Println()
}

func StartProcess(inputUri string, outfile string, args []string) bool {
	//	procAttr := &os.ProcAttr{
	//		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
	//	}
	//	userArgs := strings.Split(config.Conf.DezoomifyRs, " ")
	//	argv := []string{""}
	//	if userArgs != nil {
	//		argv = append(argv, userArgs...)
	//	}
	//	if args != nil {
	//		argv = append(argv, args...)
	//	}
	//	argv = append(argv, inputUri, outfile)
	//	process, err := os.StartProcess(config.Conf.DezoomifyPath, argv, procAttr)
	//	if err != nil {
	//		fmt.Println("start process error:", err)
	//		return false
	//	}
	//	_, err = process.Wait()
	//	if err != nil {
	//		fmt.Println("wait error:", err)
	//		return false
	//	}
	//	fmt.Println()
	return false
}

func OpenWebBrowser(args []string) bool {

	return true
}

func isProcessRunning(processName string) (bool, error) {

	return false, nil
}

func IsBookgetGuiRunning() (ok bool, err error) {
	return false, nil
}
