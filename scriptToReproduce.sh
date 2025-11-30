#!/bin/bash

# --- Inicio del script ---
set -e # El script se detendrá si un comando falla
set -u # El script se detendrá si se usa una variable no definida

# Función para mostrar ayuda
show_help() {
    echo "Uso: $0 [opciones]"
    echo ""
    echo "Opciones:"
    echo "  -f, --file FILE     Especifica el archivo de ramas (por defecto: branches.txt)"
    echo "  -m, --master RAMA   Especifica la rama master (por defecto: master)"
    echo "  -d, --dry-run       Ejecuta en modo simulación (no hace cambios reales)"
    echo "  -h, --help          Muestra esta ayuda"
    echo ""
    echo "Descripción:"
    echo "  Este script fusiona automáticamente la rama master en todas las ramas"
    echo "  especificadas en el archivo de ramas."
}

# Función para limpiar y volver a la rama original
cleanup() {
    local exit_code=$?
    if [ -n "${ORIGINAL_BRANCH:-}" ] && [ "$exit_code" -ne 0 ]; then
        echo "🔄 Volviendo a la rama original debido a error: '$ORIGINAL_BRANCH'..."
        git checkout "$ORIGINAL_BRANCH" 2>/dev/null || true
    fi
}

# Configurar trap para limpieza en caso de error o interrupción
trap cleanup INT TERM

# --- Configuración por defecto ---
BRANCH_FILE="branches.txt"
MASTER_BRANCH="master"
DRY_RUN=false

# Procesar argumentos de línea de comandos
while [ $# -gt 0 ]; do
    case $1 in
        -f|--file)
            BRANCH_FILE="$2"
            shift 2
            ;;
        -m|--master)
            MASTER_BRANCH="$2"
            shift 2
            ;;
        -d|--dry-run)
            DRY_RUN=true
            shift
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            echo "❌ Error: Opción desconocida '$1'"
            show_help
            exit 1
            ;;
    esac
done

# 1. Verificar que estamos en un repositorio git
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    echo "❌ Error: No estás en un repositorio git."
    exit 1
fi

# 2. Guardar el nombre de la rama actual
ORIGINAL_BRANCH=$(git branch --show-current)
if [ -z "$ORIGINAL_BRANCH" ]; then
    echo "❌ Error: No se pudo determinar la rama actual."
    exit 1
fi

echo "📍 Estabas en la rama '$ORIGINAL_BRANCH'. Volveremos a ella al finalizar."
if [ "$DRY_RUN" = true ]; then
    echo "🔍 MODO SIMULACIÓN: No se realizarán cambios reales."
fi
echo "---"

# 3. Comprobar si el archivo de ramas existe
if [ ! -f "$BRANCH_FILE" ]; then
    echo "❌ Error: El archivo '$BRANCH_FILE' no se encuentra."
    echo "💡 Crea un archivo con una rama por línea, por ejemplo:"
    echo "   echo 'develop' > $BRANCH_FILE"
    echo "   echo 'feature/nueva-funcionalidad' >> $BRANCH_FILE"
    exit 1
fi

# 4. Verificar que el archivo no esté vacío
if [ ! -s "$BRANCH_FILE" ]; then
    echo "❌ Error: El archivo '$BRANCH_FILE' está vacío."
    exit 1
fi

# 5. Verificar que la rama master existe
if ! git show-ref --verify --quiet "refs/heads/$MASTER_BRANCH" && ! git show-ref --verify --quiet "refs/remotes/origin/$MASTER_BRANCH"; then
    echo "❌ Error: La rama '$MASTER_BRANCH' no existe localmente ni en origin."
    exit 1
fi

# 6. Actualizar la información del repositorio remoto
echo "🔄 Actualizando la información de 'origin'..."
if [ "$DRY_RUN" = false ]; then
    git fetch origin
fi
echo "✅ Repositorio actualizado."
echo "---"

# Función para verificar si una rama existe en remoto
branch_exists_remote() {
    git ls-remote --heads origin "$1" 2>/dev/null | grep -q "refs/heads/$1"
}

# Función para verificar si hay cambios sin confirmar
has_uncommitted_changes() {
    ! git diff-index --quiet HEAD --
}

# Función para procesar una rama
process_branch() {
    local branch="$1"
    local success=true

    echo ">>> Procesando la rama: $branch"

    # Verificar si la rama existe en remoto
    if ! branch_exists_remote "$branch"; then
        echo "⚠️  Advertencia: La rama '$branch' no existe en origin. Saltando..."
        echo "💡 Puedes verificar las ramas disponibles con: git branch -r | grep $branch"
        return 0
    fi

    # Cambiar a la rama
    if [ "$DRY_RUN" = false ]; then
        if ! git checkout "$branch" 2>/dev/null; then
            echo "❌ Error: No se pudo cambiar a la rama '$branch'."
            return 1
        fi
    else
        echo "🔍 [SIMULACIÓN] git checkout $branch"
    fi

    # Verificar cambios sin confirmar
    if [ "$DRY_RUN" = false ] && has_uncommitted_changes; then
        echo "⚠️  Advertencia: Hay cambios sin confirmar en la rama '$branch'. Saltando..."
        return 1
    fi

    # --- INICIO DE LA SINCRONIZACIÓN ---
    echo "📥 Trayendo cambios de origin/$branch..."
    if [ "$DRY_RUN" = false ]; then
        if ! git pull origin "$branch"; then
            echo "❌ Error: No se pudieron traer los cambios de origin/$branch."
            return 1
        fi
    else
        echo "� [SIMULACIÓN] git pull origin $branch"
    fi

    echo "�🔄 Fusionando origin/$MASTER_BRANCH en $branch..."
    if [ "$DRY_RUN" = false ]; then
        if ! git merge "origin/$MASTER_BRANCH" --no-ff -m "Merge origin/$MASTER_BRANCH into $branch"; then
            echo "❌ Error: Conflicto al fusionar origin/$MASTER_BRANCH en $branch."
            echo "💡 Resuelve los conflictos manualmente y ejecuta el script nuevamente."
            return 1
        fi
    else
        echo "🔍 [SIMULACIÓN] git merge origin/$MASTER_BRANCH --no-ff -m \"Merge origin/$MASTER_BRANCH into $branch\""
    fi

    echo "⬆️ Subiendo cambios a origin/$branch..."
    if [ "$DRY_RUN" = false ]; then
        if ! git push origin "$branch"; then
            echo "❌ Error: No se pudieron subir los cambios a origin/$branch."
            return 1
        fi
    else
        echo "🔍 [SIMULACIÓN] git push origin $branch"
    fi
    # --- FIN DE LA SINCRONIZACIÓN ---

    echo "✅ Sincronización completada para la rama: $branch"
    echo "---"
    return 0
}

# 7. Leer el archivo y procesar cada rama
FAILED_BRANCHES=()
PROCESSED_COUNT=0
# Contar líneas no vacías y que no sean comentarios
TOTAL_BRANCHES=$(grep -v '^[[:space:]]*$' "$BRANCH_FILE" | grep -v '^[[:space:]]*#' | wc -l)

echo "📊 Se procesarán $TOTAL_BRANCHES ramas desde '$BRANCH_FILE'"
echo "🎯 Rama master: '$MASTER_BRANCH'"
echo "---"

while IFS= read -r branch; do
    # Saltar líneas vacías y comentarios
    if [[ -z "$branch" || "$branch" =~ ^[[:space:]]*# ]]; then
        continue
    fi

    # Eliminar espacios en blanco al inicio y final
    branch=$(echo "$branch" | xargs)

    if [ -n "$branch" ]; then
        PROCESSED_COUNT=$((PROCESSED_COUNT + 1))
        echo "[$PROCESSED_COUNT/$TOTAL_BRANCHES] Procesando: $branch"

        # Usar set +e temporalmente para manejar errores de ramas no existentes
        set +e
        if ! process_branch "$branch"; then
            FAILED_BRANCHES+=("$branch")
        fi
        set -e
    fi
done < "$BRANCH_FILE"

# 8. Mostrar resumen final
echo "🎉 Proceso finalizado."
echo "📊 Resumen:"
echo "   - Ramas procesadas: $PROCESSED_COUNT"
echo "   - Ramas exitosas: $((PROCESSED_COUNT - ${#FAILED_BRANCHES[@]}))"
echo "   - Ramas fallidas: ${#FAILED_BRANCHES[@]}"

if [ ${#FAILED_BRANCHES[@]} -gt 0 ]; then
    echo ""
    echo "❌ Ramas que fallaron:"
    for failed_branch in "${FAILED_BRANCHES[@]}"; do
        echo "   - $failed_branch"
    done
    echo ""
    echo "💡 Revisa los errores anteriores y ejecuta el script nuevamente para las ramas fallidas."
fi

# Volver a la rama original al finalizar exitosamente
if [ -n "${ORIGINAL_BRANCH:-}" ]; then
    echo ""
    echo "↩️ Volviendo a tu rama original: '$ORIGINAL_BRANCH'..."
    git checkout "$ORIGINAL_BRANCH" 2>/dev/null || true
fi