package main

import (
	"fmt"

	filter "github.com/antlinker/go-dirtyfilter"
	"github.com/antlinker/go-dirtyfilter/store"
)

var (
	filterText = `你是大傻子`
)

func main() {
	memStore, err := store.NewMemoryStore(store.MemoryConfig{
		DataSource: []string{"傻子", "坏蛋", "傻缺", "傻屌", "傻大个"},
	})
	if err != nil {
		panic(err)
	}
	filterManage := filter.NewDirtyManager(memStore)
	result, err := filterManage.Filter().Replace(filterText, '*')
	if err != nil {
		panic(err)
	}
	fmt.Println(result)
}
