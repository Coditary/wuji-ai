package clix

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coditary/wuji-core/pkg/driver"
)

type RAGOpFlags struct {
	Index           bool
	Query           bool
	Answer          bool
	ListCollections bool
	Info            bool
	Stats           bool
	DeleteCol       bool
	Purge           bool
	Export          bool
	ImportCol       bool
	Rename          bool
}

func (f RAGOpFlags) Resolve() (driver.RAGTask, error) {
	var task driver.RAGTask
	count := 0
	pick := func(set bool, t driver.RAGTask) error {
		if !set {
			return nil
		}
		count++
		if count > 1 {
			return fmt.Errorf("choose one rag operation flag (--index, --query, --answer, …)")
		}
		task = t
		return nil
	}
	checks := []struct {
		set  bool
		task driver.RAGTask
	}{
		{f.Index, driver.RAGTaskIndex},
		{f.Query, driver.RAGTaskQuery},
		{f.Answer, driver.RAGTaskAnswer},
		{f.ListCollections, driver.RAGTaskList},
		{f.Info, driver.RAGTaskInfo},
		{f.Stats, driver.RAGTaskStats},
		{f.DeleteCol, driver.RAGTaskDelete},
		{f.Purge, driver.RAGTaskPurge},
		{f.Export, driver.RAGTaskExport},
		{f.ImportCol, driver.RAGTaskImport},
		{f.Rename, driver.RAGTaskRename},
	}
	for _, c := range checks {
		if err := pick(c.set, c.task); err != nil {
			return "", err
		}
	}
	if task == "" {
		return "", fmt.Errorf("one operation flag is required (--index, --query, --answer, --list-collections, …)")
	}
	return task, nil
}

func AddRAGOpFlags(cmd *cobra.Command, f *RAGOpFlags) {
	cmd.Flags().BoolVar(&f.Index, "index", false, "index documents into a collection")
	cmd.Flags().BoolVar(&f.Query, "query", false, "retrieve relevant chunks")
	cmd.Flags().BoolVar(&f.Answer, "answer", false, "retrieve chunks and generate an answer")
	cmd.Flags().BoolVar(&f.ListCollections, "list-collections", false, "list indexed collections (prefer: wuji list collections)")
	cmd.Flags().BoolVar(&f.Info, "info", false, "show collection metadata (prefer: wuji info collection <name>)")
	cmd.Flags().BoolVar(&f.Stats, "stats", false, "show detailed collection statistics")
	cmd.Flags().BoolVar(&f.DeleteCol, "delete", false, "delete a collection")
	cmd.Flags().BoolVar(&f.Purge, "purge", false, "remove one source from a collection")
	cmd.Flags().BoolVar(&f.Export, "export", false, "export a collection directory snapshot")
	cmd.Flags().BoolVar(&f.ImportCol, "import", false, "import a collection directory snapshot")
	cmd.Flags().BoolVar(&f.Rename, "rename", false, "rename a collection")
}
