package controller

import (
	appsync "autoMerger/internal/application/sync"
	"autoMerger/internal/interfaces/cli"
	"fmt"
	"os"
	"os/exec"
	"sync"
)

type SyncController struct {
    service *appsync.SyncService
}

func NewSyncController(service *appsync.SyncService) *SyncController {
    return &SyncController{
        service: service,
    }
}

func (s *SyncController) Run(args cli.CLIArgs) error {

	if err := s.service.CheckGitRepo(args.Path); err != nil {
		return fmt.Errorf("la siguiente ruta no contiene ningun repositorio git: %s", args.Path)
	}

	if !args.DryRun {
		
		cmd := exec.Command("git", "fetch", "origin")
		cmd.Dir = args.Path
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("error actualizando origin: %w", err)
		}
	}

	if err := s.service.CheckoutMaster(args.Master, args.Path); err != nil {	
		return fmt.Errorf("error al cambiar a master: %w", err)
	}

	if err := s.service.LoadBranches(args.BranchFile); err != nil{
		println("Error leyendo ramas")
		return fmt.Errorf("error cargando el archivo de las ramas: %w", err)
	}

	appsync.CleanAllWorktrees(args.Path)

	appsync.CreateWorkTrees(&s.service.Branches, args.Path)

	var wg sync.WaitGroup

	for _, branch := range s.service.Branches {
		wg.Add(1)
		branch := branch // importante: crear copia local para la goroutine
		go func() {
			defer wg.Done()
			if err := s.service.SyncBranch(branch, args.Master, args.Path , args.DryRun); err != nil {
				fmt.Printf("%s -> error: %s \n", branch, err)
			}
		}()
	}


	wg.Wait()

	appsync.CleanAllWorktrees(args.Path)

	if err := s.service.RestoreOriginalBranch(args.Path); err != nil {
		return fmt.Errorf("error al restaurar la rama original: %w", err)
	}

	return nil

}
