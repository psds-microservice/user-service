//go:build ignore
// +build ignore

// gen_rawdesc reads fd.pb (FileDescriptorSet) and prints the first file's
// serialized FileDescriptorProto as a Go string literal for use in user_service.pb.go.
// Run from repo root: go run scripts/gen_rawdesc.go
package main

import (
	"fmt"
	"os"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

func main() {
	b, err := os.ReadFile("fd.pb")
	if err != nil {
		fmt.Fprintf(os.Stderr, "read fd.pb: %v\n", err)
		os.Exit(1)
	}
	var set descriptorpb.FileDescriptorSet
	if err := proto.Unmarshal(b, &set); err != nil {
		fmt.Fprintf(os.Stderr, "unmarshal: %v\n", err)
		os.Exit(1)
	}
	if len(set.File) == 0 {
		fmt.Fprintf(os.Stderr, "no file in set\n")
		os.Exit(1)
	}
	single, err := proto.Marshal(set.File[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "marshal: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("%q", single)
}
