// 假设该文件存在，若不存在需要创建
package main

import (
	"flag"
	"fmt"
	"hollow/cmd/hollow_cli/generator"
)

func main() {
	protoPath := flag.String("proto", "", "Path to the proto file")
	generateHTTP := flag.Bool("http", false, "Generate HTTP handler code")
	generateGo := flag.Bool("go", false, "Generate Go code from proto file")
	flag.Parse()

	if *protoPath == "" {
		fmt.Println("Please provide a proto file path using -proto flag")
		return
	}

	if *generateHTTP {
		err := generator.GenerateProto(*protoPath)
		if err != nil {
			fmt.Printf("Failed to generate HTTP handler code: %v\n", err)
		} else {
			fmt.Println("HTTP handler code generated successfully")
		}
	}

	if *generateGo {
		err := generator.GenerateGoFromProto(*protoPath)
		if err != nil {
			fmt.Printf("Failed to generate Go code from proto file: %v\n", err)
		} else {
			fmt.Println("Go code generated successfully")
		}
	}
}
