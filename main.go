//go:build linux

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/frida/frida-go/frida"
)

// 定义命令行参数
var (
	pidFlag      = flag.Int("pid", 0, "目标进程 PID")
	injectFile   = flag.String("inject-file", "", "要注入的共享对象文件路径")
	entrypoint   = flag.String("entrypoint", "example_main", "要调用的入口函数名称")
	data         = flag.String("data", "w00t", "传递给入口函数的数据")
	spawnProgram = flag.String("spawn", "", "如果不使用 PID，可以指定要启动的程序")
	verbose      = flag.Bool("verbose", false, "启用详细日志")
)

func usage() {
	fmt.Fprintf(os.Stderr, "用法: %s [选项]\n\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "选项:\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\n示例:\n")
	fmt.Fprintf(os.Stderr, "  # 向已运行的进程注入库\n")
	fmt.Fprintf(os.Stderr, "  %s --pid 12345 --inject-file ./hook_open.so\n\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  # 使用 spawn-gating 启动并注入程序\n")
	fmt.Fprintf(os.Stderr, "  %s --spawn /usr/bin/top --inject-file ./hook_open.so\n", os.Args[0])
	os.Exit(1)
}

func main() {
	// 配置日志
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	log.Println("Frida 注入工具启动")

	// 解析命令行参数
	flag.Usage = usage
	flag.Parse()

	// 验证参数有效性
	if *injectFile == "" {
		log.Println("错误: 必须指定要注入的文件 (--inject-file)")
		flag.Usage()
	}

	// 检查注入文件是否存在
	absInjectPath, err := filepath.Abs(*injectFile)
	if err != nil {
		log.Fatalf("无法获取注入文件的绝对路径: %v", err)
	}
	if _, err := os.Stat(absInjectPath); os.IsNotExist(err) {
		log.Fatalf("注入文件不存在: %s", absInjectPath)
	}

	// 检查注入方式
	if *pidFlag == 0 && *spawnProgram == "" {
		log.Println("错误: 必须指定目标进程 PID (--pid) 或要启动的程序 (--spawn)")
		flag.Usage()
	}

	// 初始化 Frida
	mgr := frida.NewDeviceManager()
	if *verbose {
		log.Println("正在枚举 Frida 设备...")
	}

	devices, err := mgr.EnumerateDevices()
	if err != nil {
		log.Fatalf("枚举 Frida 设备失败: %v", err)
	}

	if *verbose {
		log.Printf("发现 %d 个 Frida 设备", len(devices))
		for _, d := range devices {
			log.Printf("[*] Found device with id: %s", d.ID())
		}
	}

	localDev, err := mgr.LocalDevice()
	if err != nil {
		log.Fatalf("无法获取本地 Frida 设备: %v", err)
	}
	log.Printf("[*] 使用设备: %s", localDev.Name())

	// 类型断言转换为所需类型
	localDevPtr := localDev.(*frida.Device)

	// 根据参数进行操作
	if *spawnProgram != "" {
		// 使用 spawn-gating
		spawnAndInject(localDevPtr, *spawnProgram, absInjectPath, *entrypoint, *data, *verbose)
	} else {
		// 直接注入到现有进程
		injectToProcess(localDevPtr, *pidFlag, absInjectPath, *entrypoint, *data)
	}
}

// injectToProcess 直接向现有进程注入库
func injectToProcess(device *frida.Device, pid int, libPath string, entrypoint string, data string) {
	log.Printf("尝试将 Frida 库注入进程 PID=%d, 路径=%s", pid, libPath)
	_, err := device.InjectLibraryFile(pid, libPath, entrypoint, data)
	if err != nil {
		log.Fatalf("注入库失败: %s", err)
		return
	}
	log.Printf("注入成功! 库已加载到进程 PID=%d", pid)
}

// spawnAndInject 使用 spawn-gating 启动程序并注入库
func spawnAndInject(device *frida.Device, programPath string, libPath string, entrypoint string, data string, verbose bool) {
	if verbose {
		log.Printf("使用 spawn-gating 启动并注入程序: %s", programPath)
	}

	// 使用 spawn 创建进程
	pid, err := device.Spawn(programPath, nil)
	if err != nil {
		log.Fatalf("无法启动进程 %s: %v", programPath, err)
		return
	}
	log.Printf("进程已创建, PID=%d", pid)

	// 在进程启动但尚未执行前注入
	_, err = device.InjectLibraryFile(pid, libPath, entrypoint, data)
	if err != nil {
		log.Printf("注入库失败: %s", err)
		return
	}
	log.Printf("注入成功! 库已加载到进程 PID=%d", pid)

	// 恢复进程执行
	err = device.Resume(pid)
	if err != nil {
		log.Fatalf("无法恢复进程: %v", err)
		return
	}
	log.Printf("进程已恢复执行")

	log.Printf("注意: 主程序不会等待目标进程结束。您可能需要按 Ctrl+C 退出此注入器。")
}
