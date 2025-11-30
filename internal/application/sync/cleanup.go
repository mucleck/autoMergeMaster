package sync

import (
    "fmt"
    "os"
    "path/filepath"
)

func CleanAllWorktrees(repoPath string) {

    worktreesRoot := filepath.Join(repoPath, ".automerger")
    if _, err := os.Stat(worktreesRoot); err == nil {
        os.RemoveAll(worktreesRoot)
    }

    gitWorktrees := filepath.Join(repoPath, ".git", "worktrees")
    entries, err := os.ReadDir(gitWorktrees)
    if err == nil {
        for _, e := range entries {
            p := filepath.Join(gitWorktrees, e.Name())
            os.RemoveAll(p)
        }
    }

    fmt.Println("Limpieza global de worktrees completada.")
}
