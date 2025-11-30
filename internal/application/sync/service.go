package sync

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type SyncService struct {
	Branches []string
	OriginalBranch string
}

func NewSyncService() *SyncService {
	return &SyncService{}
}

func (s *SyncService) CheckGitRepo(path string) error {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = path
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("no estás en un repositorio git")
	}
	return nil
}


//Devuelvo array con nombres de las ramas parseados
func (s *SyncService) LoadBranches(file string) error {

	content, err := os.ReadFile(file)
    if err != nil {
        return err
    }

    lines := strings.Split(string(content), "\n")

    for _, line := range lines {
        line = strings.TrimSpace(line)
        if line == "" || strings.HasPrefix(line, "#") {
            continue
        }
        s.Branches = append(s.Branches, line)
    }

    
    return nil

}

func (s *SyncService) SyncBranch(branch, master, repoPath string, dryRun bool) error {
	

    worktreesRoot := filepath.Join(repoPath, ".automerger")
    worktreePath := filepath.Join(worktreesRoot, branch)

    if dryRun {
        /*fmt.Printf("[SIMULACIÓN] Crear worktree %s\n", worktreePath)
        fmt.Printf("[SIMULACIÓN] Pull origin/%s\n", branch)
        fmt.Printf("[SIMULACIÓN] Merge origin/%s → %s\n", master, branch)
        fmt.Printf("[SIMULACIÓN] Push %s\n", branch)*/
        return nil
    }

    // Hacer pull
    cmd := exec.Command("git", "pull", "origin", branch)
    cmd.Dir = worktreePath
	cmd.Stdout = nil
	cmd.Stderr = nil
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("%s -> error haciendo pull: %w", branch, err)
    }

    // Merge
    cmd = exec.Command(
        "git", "merge", "--no-ff", "-m",
        fmt.Sprintf("Merge origin/%s into %s", master, branch),
        "origin/"+master,
    )
    cmd.Dir = worktreePath
	cmd.Stdout = nil
	cmd.Stderr = nil
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("error mergeando %s en %s: %w", master, branch, err)
    }

    // Push
    cmd = exec.Command("git", "push", "origin", branch)
    cmd.Dir = worktreePath
	cmd.Stdout = nil
	cmd.Stderr = nil
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("error haciendo push de %s: %w", branch, err)
    }

    fmt.Printf("Rama %s sincronizada correctamente\n", branch)
    return nil

}

// Cambiar a master guardando la rama actual
func (s *SyncService) CheckoutMaster(master string, repoPath string) error {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = repoPath
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("no se pudo obtener la rama actual: %w", err)
	}
	s.OriginalBranch = string(out)
	s.OriginalBranch = s.OriginalBranch[:len(s.OriginalBranch)-1] // quitar newline

	// Cambiar a master
	cmd = exec.Command("git", "checkout", master)
	cmd.Dir = repoPath
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("no se pudo cambiar a master: %w", err)
	}

	return nil
}

// Volver a la rama original guardada
func (s *SyncService) RestoreOriginalBranch(repoPath string) error {
	if s.OriginalBranch == "" {
		return fmt.Errorf("no se ha guardado la rama original")
	}

	cmd := exec.Command("git", "checkout", s.OriginalBranch)
	cmd.Dir = repoPath
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("no se pudo volver a la rama original %s: %w", s.OriginalBranch, err)
	}

	return nil
}

func CreateWorkTrees(branches *[]string, repoPath string){

    worktreesRoot := filepath.Join(repoPath, ".automerger")


    if err := os.MkdirAll(worktreesRoot, 0755); err != nil {
        fmt.Printf("-> no se pudo crear el directorio de worktrees: %w", err)
    }

    for _, branch := range *branches {

        worktreePath := filepath.Join(worktreesRoot, branch)

        if _, err := os.Stat(worktreePath); err == nil {
            os.RemoveAll(worktreePath)
        }

        // Crear el worktree
        cmd := exec.Command("git", "worktree", "add", filepath.Join(worktreesRoot, branch), branch)
        //"origin/"+branch, "-b", branch
        cmd.Dir = repoPath
        if err := cmd.Run(); err != nil {
           fmt.Printf("%s -> error creando worktree: %w", branch, err)
        }
    }

    fmt.Println("Creados todos los worktrees con exito")
}