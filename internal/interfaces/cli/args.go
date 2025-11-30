package cli

import (
	"flag"
	"fmt"
)

/* Ejemplo de como esta en el script bash esto:

    echo "  -f, --file FILE     Especifica el archivo de ramas (por defecto: branches.txt)"
    echo "  -m, --master RAMA   Especifica la rama master (por defecto: master)"
    echo "  -d, --dry-run       Ejecuta en modo simulación (no hace cambios reales)"
    echo "  -h, --help          Muestra esta ayuda"

*/
type CLIArgs struct {
	BranchFile string
    Master     string
	Path string
    DryRun     bool
    ShowHelp   bool
}

func ParseArgs(input []string) CLIArgs {
    var args CLIArgs

    // Crear un conjunto de flags independiente
    fs := flag.NewFlagSet("sync-branches", flag.ContinueOnError)

    fs.StringVar(&args.BranchFile, "f", "branches.txt", "Archivo con las ramas")
    fs.StringVar(&args.BranchFile, "file", "branches.txt", "Archivo con las ramas")

    fs.StringVar(&args.Master, "m", "master", "Rama master")
    fs.StringVar(&args.Master, "master", "master", "Rama master")

	fs.StringVar(&args.Path, "p", ".", "Define el path del repositorio")

    fs.BoolVar(&args.DryRun, "d", false, "Modo simulación")
    fs.BoolVar(&args.DryRun, "dry-run", false, "Modo simulación")

    fs.BoolVar(&args.ShowHelp, "h", false, "Mostrar ayuda")
    fs.BoolVar(&args.ShowHelp, "help", false, "Mostrar ayuda")

    fs.Parse(input)

    return args
}

func PrintHelp() {
    fmt.Println("Uso:")
    fmt.Println("  -f, --file FILE     Archivo con las ramas (default: branches.txt)")
    fmt.Println("  -m, --master RAMA   Rama master (default: master)")
    fmt.Println("  -d, --dry-run       Modo simulación (no hace cambios)")
    fmt.Println("  -p,                 Ruta absoluta del repo (En blanco para usar la ruta actual)")
    fmt.Println("  -h, --help          Muestra esta ayuda")
}
