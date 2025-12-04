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
    os.RemoveAll(gitWorktrees)

    fmt.Println("Limpieza global de worktrees completada.")
}
