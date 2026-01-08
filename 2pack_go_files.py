#!/usr/bin/env python3
"""
Оптимизированный скрипт для упаковки Go-проекта в формат,
удобный для языковых моделей (Claude, GPT и др.)

Формат вывода — плоский текст с XML-подобными разделителями:
- Легко парсится LLM
- Код остаётся читаемым без экранирования
- Минимальный оверхед по токенам
"""

import os
import sys
import argparse
from pathlib import Path
from typing import List, Set, Optional


# ═══════════════════════════════════════════════════════════════════
#   КОНФИГУРАЦИЯ
# ═══════════════════════════════════════════════════════════════════

DEFAULT_EXTENSIONS = {".go", ".mod", ".sum", ".json", ".yaml", ".yml", ".md", ".sql", ".proto"}

DEFAULT_IGNORE = {
    # Системные
    ".git", ".svn", ".hg",
    ".idea", ".vscode", ".kilocode",
    
    # Зависимости
    "vendor", "node_modules", ".venv",

    # Игнорируемые директории
    "mcp-server",
    
    # Сборка и кэш
    "bin", "dist", "build", "_build",
    "ftp_cache", "tmp", ".cache",
    
    # Тесты и моки (опционально — можно убрать)
    # "testdata", "mock", "mocks",
    
    # Сам скрипт
    "2pack_go_files.py", "project_dump.py", "project_dump.txt",
}

# Файлы, которые важны и должны идти первыми
PRIORITY_FILES = ["go.mod", "go.sum", "README.md", "main.go", "config.yaml", "config.json"]


# ═══════════════════════════════════════════════════════════════════
#   ЛОГИКА СБОРА ФАЙЛОВ
# ═══════════════════════════════════════════════════════════════════

def should_ignore(path: Path, ignore_set: Set[str]) -> bool:
    """Проверяет, нужно ли игнорировать путь."""
    parts = path.parts
    for part in parts:
        if part in ignore_set:
            return True
        # Игнорируем скрытые файлы/папки (начинаются с точки), кроме важных
        if part.startswith(".") and part not in {".env", ".env.example", ".gitignore"}:
            if part in ignore_set or part.startswith("."):
                # Проверяем, не в явном ли списке игнора
                pass
    
    for ignore in ignore_set:
        if ignore in str(path):
            return True
    return False


def collect_files(root: Path, extensions: Set[str], ignore_set: Set[str]) -> List[Path]:
    """Собирает все файлы с нужными расширениями."""
    files = []
    
    for path in root.rglob("*"):
        if path.is_file() and path.suffix in extensions:
            if not should_ignore(path.relative_to(root), ignore_set):
                files.append(path)
    
    return files


def sort_files(files: List[Path], root: Path) -> List[Path]:
    """
    Сортирует файлы для оптимального чтения LLM:
    1. Приоритетные файлы (go.mod, README, main.go)
    2. Корневые файлы
    3. cmd/ — точки входа
    4. internal/ — внутренняя логика
    5. pkg/ — переиспользуемый код
    6. Остальное по алфавиту
    """
    
    def sort_key(path: Path) -> tuple:
        rel = path.relative_to(root)
        name = path.name
        parts = rel.parts
        
        # Приоритет по имени файла
        if name in PRIORITY_FILES:
            priority = PRIORITY_FILES.index(name)
        else:
            priority = 100
        
        # Приоритет по директории
        if len(parts) == 1:
            dir_priority = 0  # Корень
        elif parts[0] == "cmd":
            dir_priority = 1
        elif parts[0] == "internal":
            dir_priority = 2
        elif parts[0] == "pkg":
            dir_priority = 3
        elif parts[0] == "api":
            dir_priority = 4
        else:
            dir_priority = 10
        
        return (dir_priority, priority, str(rel))
    
    return sorted(files, key=sort_key)


def build_tree(root: Path, ignore_set: Set[str], extensions: Set[str]) -> str:
    """Строит компактное дерево проекта."""
    lines = ["."]
    
    def walk(dir_path: Path, prefix: str = ""):
        try:
            entries = sorted(dir_path.iterdir(), key=lambda p: (not p.is_dir(), p.name.lower()))
        except PermissionError:
            return
        
        # Фильтруем
        entries = [
            e for e in entries 
            if not should_ignore(e.relative_to(root), ignore_set)
            and (e.is_dir() or e.suffix in extensions)
        ]
        
        for idx, entry in enumerate(entries):
            is_last = idx == len(entries) - 1
            connector = "└── " if is_last else "├── "
            
            display_name = entry.name
            if entry.is_dir():
                display_name += "/"
            
            lines.append(f"{prefix}{connector}{display_name}")
            
            if entry.is_dir():
                new_prefix = prefix + ("    " if is_last else "│   ")
                walk(entry, new_prefix)
    
    walk(root)
    return "\n".join(lines)


# ═══════════════════════════════════════════════════════════════════
#   ФОРМАТИРОВАНИЕ ВЫВОДА
# ═══════════════════════════════════════════════════════════════════

def format_output(
    tree: str,
    files: List[Path],
    root: Path,
    include_stats: bool = True
) -> str:
    """Форматирует финальный вывод."""
    
    parts = []
    
    # Заголовок с метаданными
    if include_stats:
        total_lines = 0
        total_size = 0
        for f in files:
            try:
                content = f.read_text(encoding="utf-8", errors="ignore")
                total_lines += content.count("\n") + 1
                total_size += len(content)
            except:
                pass
        
        parts.append(f"<!-- Project: {root.resolve().name} -->")
        parts.append(f"<!-- Files: {len(files)}, Lines: {total_lines}, Size: {total_size // 1024}KB -->")
        parts.append("")
    
    # Дерево проекта
    parts.append("<tree>")
    parts.append(tree)
    parts.append("</tree>")
    parts.append("")
    
    # Файлы
    for file_path in files:
        rel_path = file_path.relative_to(root)
        
        try:
            content = file_path.read_text(encoding="utf-8", errors="ignore")
        except Exception as e:
            content = f"<!-- Error reading file: {e} -->"
        
        # Убираем trailing whitespace и лишние пустые строки в конце
        content = content.rstrip()
        
        parts.append(f'<file path="{rel_path}">')
        parts.append(content)
        parts.append("</file>")
        parts.append("")
    
    return "\n".join(parts)


# ═══════════════════════════════════════════════════════════════════
#   CLI
# ═══════════════════════════════════════════════════════════════════

def main():
    parser = argparse.ArgumentParser(
        description="Упаковка Go-проекта в формат для LLM",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Примеры:
  %(prog)s                           # Упаковать текущую директорию
  %(prog)s -o context.txt            # Указать имя выходного файла  
  %(prog)s -e .go .proto             # Только .go и .proto файлы
  %(prog)s --ignore vendor testdata  # Дополнительно игнорировать
  %(prog)s --stdout                  # Вывести в stdout (для пайпов)
        """
    )
    
    parser.add_argument(
        "root", 
        nargs="?", 
        default=".",
        help="Корневая директория проекта (по умолчанию: .)"
    )
    
    parser.add_argument(
        "-o", "--output",
        default="project_context.txt",
        help="Выходной файл (по умолчанию: project_context.txt)"
    )
    
    parser.add_argument(
        "-e", "--extensions",
        nargs="+",
        help=f"Расширения файлов (по умолчанию: {' '.join(sorted(DEFAULT_EXTENSIONS))})"
    )
    
    parser.add_argument(
        "--ignore",
        nargs="+",
        default=[],
        help="Дополнительные пути для игнорирования"
    )
    
    parser.add_argument(
        "--stdout",
        action="store_true",
        help="Вывести результат в stdout вместо файла"
    )
    
    parser.add_argument(
        "--no-stats",
        action="store_true",
        help="Не включать статистику в вывод"
    )
    
    args = parser.parse_args()
    
    root = Path(args.root).resolve()
    if not root.is_dir():
        print(f"Ошибка: {root} не является директорией", file=sys.stderr)
        sys.exit(1)
    
    # Собираем конфиг
    extensions = set(args.extensions) if args.extensions else DEFAULT_EXTENSIONS
    ignore_set = DEFAULT_IGNORE | set(args.ignore) | {args.output}
    
    # Собираем и сортируем файлы
    files = collect_files(root, extensions, ignore_set)
    files = sort_files(files, root)
    
    if not files:
        print("Предупреждение: не найдено файлов для упаковки", file=sys.stderr)
    
    # Строим дерево
    tree = build_tree(root, ignore_set, extensions)
    
    # Форматируем вывод
    output = format_output(tree, files, root, include_stats=not args.no_stats)
    
    # Выводим
    if args.stdout:
        print(output)
    else:
        output_path = Path(args.output)
        output_path.write_text(output, encoding="utf-8")
        print(f"✓ Упаковано {len(files)} файлов → {output_path}", file=sys.stderr)
        print(f"  Размер: {len(output) // 1024}KB ({len(output)} байт)", file=sys.stderr)


if __name__ == "__main__":
    main()