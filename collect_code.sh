

# Имя выходного файла
OUTPUT_FILE="project_context.txt"

# Очищаем выходной файл, если он существует
> "$OUTPUT_FILE"

echo "Collecting project files into $OUTPUT_FILE..."

# Функция для добавления файла в отчет
add_file() {
    local file="$1"
    if [ -f "$file" ]; then
        echo "================================================================================" >> "$OUTPUT_FILE"
        echo "FILE: $file" >> "$OUTPUT_FILE"
        echo "================================================================================" >> "$OUTPUT_FILE"
        cat "$file" >> "$OUTPUT_FILE"
        echo -e "\n\n" >> "$OUTPUT_FILE"
        echo "Added: $file"
    else
        echo "Warning: File not found - $file"
    fi
}

# 1. Добавляем документацию
add_file "README.md"
add_file "CLAUDE.md"
add_file "go.mod"

# 2. Добавляем конфигурацию
add_file "configs/config.yaml"

# 3. Собираем Go файлы по ключевым директориям

# Main / Entry point
find cmd -name "*.go" | sort | while read -r file; do
    add_file "$file"
done

# Internal packages
# Используем find для рекурсивного поиска, исключая тесты (_test.go), 
# чтобы не засорять контекст, если они не нужны.
# Если тесты нужны, убери 'grep -v "_test.go"'

echo "Processing internal packages..."

find internal -name "*.go" | grep -v "_test.go" | sort | while read -r file; do
    add_file "$file"
done

find pkg -name "*.go" | grep -v "_test.go" | sort | while read -r file; do
    add_file "$file"
done

echo "Done! All code collected in $OUTPUT_FILE"
