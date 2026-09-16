package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coditary/wuji-core/pkg/driver"
)

type ragOpFlags struct {
	index           bool
	query           bool
	answer          bool
	listCollections bool
	info            bool
	stats           bool
	deleteCol       bool
	purge           bool
	export          bool
	importCol       bool
	rename          bool
}

func (f ragOpFlags) resolve() (driver.RAGTask, error) {
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
		{f.index, driver.RAGTaskIndex},
		{f.query, driver.RAGTaskQuery},
		{f.answer, driver.RAGTaskAnswer},
		{f.listCollections, driver.RAGTaskList},
		{f.info, driver.RAGTaskInfo},
		{f.stats, driver.RAGTaskStats},
		{f.deleteCol, driver.RAGTaskDelete},
		{f.purge, driver.RAGTaskPurge},
		{f.export, driver.RAGTaskExport},
		{f.importCol, driver.RAGTaskImport},
		{f.rename, driver.RAGTaskRename},
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

func addRAGOpFlags(cmd *cobra.Command, f *ragOpFlags) {
	cmd.Flags().BoolVar(&f.index, "index", false, "index documents into a collection")
	cmd.Flags().BoolVar(&f.query, "query", false, "retrieve relevant chunks")
	cmd.Flags().BoolVar(&f.answer, "answer", false, "retrieve chunks and generate an answer")
	cmd.Flags().BoolVar(&f.listCollections, "list-collections", false, "list indexed collections (prefer: wuji list collections)")
	cmd.Flags().BoolVar(&f.info, "info", false, "show collection metadata (prefer: wuji info collection <name>)")
	cmd.Flags().BoolVar(&f.stats, "stats", false, "show detailed collection statistics")
	cmd.Flags().BoolVar(&f.deleteCol, "delete", false, "delete a collection")
	cmd.Flags().BoolVar(&f.purge, "purge", false, "remove one source from a collection")
	cmd.Flags().BoolVar(&f.export, "export", false, "export a collection directory snapshot")
	cmd.Flags().BoolVar(&f.importCol, "import", false, "import a collection directory snapshot")
	cmd.Flags().BoolVar(&f.rename, "rename", false, "rename a collection")
}
