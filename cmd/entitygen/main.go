package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type field struct {
	Column  string
	GoName  string
	SQLType string
	GoType  string
}

var typeMap = map[string]struct {
	sql string
	goT string
}{
	"string": {sql: "TEXT", goT: "string"},
	"int":    {sql: "BIGINT", goT: "int64"},
	"bool":   {sql: "BOOLEAN", goT: "bool"},
	"time":   {sql: "TIMESTAMPTZ", goT: "time.Time"},
}

func main() {
	var name string
	var table string
	var fieldsRaw string
	flag.StringVar(&name, "name", "", "Entity name in PascalCase, example: Comment")
	flag.StringVar(&table, "table", "", "Table name in snake_case (optional)")
	flag.StringVar(&fieldsRaw, "fields", "", "Comma-separated fields, example: text:string,done:bool")
	flag.Parse()

	if strings.TrimSpace(name) == "" || strings.TrimSpace(fieldsRaw) == "" {
		exitWithError(errors.New("usage: go run ./cmd/entitygen --name Comment --fields text:string,done:bool"))
	}
	if table == "" {
		table = toSnake(name) + "s"
	}

	fields, err := parseFields(fieldsRaw)
	if err != nil {
		exitWithError(err)
	}

	migrationPath, err := createMigration(table, fields)
	if err != nil {
		exitWithError(err)
	}

	structPath, err := createStruct(name, fields)
	if err != nil {
		exitWithError(err)
	}

	fmt.Printf("created migration: %s\n", migrationPath)
	fmt.Printf("created entity: %s\n", structPath)
}

func createMigration(table string, fields []field) (string, error) {
	version, err := nextMigrationVersion(filepath.Join("internal", "db", "migrations"))
	if err != nil {
		return "", err
	}
	fileName := fmt.Sprintf("%03d_create_%s.sql", version, table)
	path := filepath.Join("internal", "db", "migrations", fileName)

	var b strings.Builder
	b.WriteString("CREATE TABLE IF NOT EXISTS " + table + " (\n")
	b.WriteString("    id BIGSERIAL PRIMARY KEY,\n")
	for _, f := range fields {
		b.WriteString(fmt.Sprintf("    %s %s NOT NULL,\n", f.Column, f.SQLType))
	}
	b.WriteString("    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),\n")
	b.WriteString("    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()\n")
	b.WriteString(");\n")

	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		return "", fmt.Errorf("write migration: %w", err)
	}
	return path, nil
}

func createStruct(name string, fields []field) (string, error) {
	path := filepath.Join("internal", "generated", toSnake(name)+"_entity.go")

	var b strings.Builder
	b.WriteString("package generated\n\n")
	b.WriteString("import \"time\"\n\n")
	b.WriteString("type " + name + " struct {\n")
	b.WriteString("\tID int64 `json:\"id\" db:\"id\"`\n")
	for _, f := range fields {
		b.WriteString(fmt.Sprintf("\t%s %s `json:\"%s\" db:\"%s\"`\n", f.GoName, f.GoType, f.Column, f.Column))
	}
	b.WriteString("\tCreatedAt time.Time `json:\"createdAt\" db:\"created_at\"`\n")
	b.WriteString("\tUpdatedAt time.Time `json:\"updatedAt\" db:\"updated_at\"`\n")
	b.WriteString("}\n")

	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		return "", fmt.Errorf("write entity struct: %w", err)
	}
	return path, nil
}

func parseFields(raw string) ([]field, error) {
	parts := strings.Split(raw, ",")
	res := make([]field, 0, len(parts))
	for _, p := range parts {
		pair := strings.Split(strings.TrimSpace(p), ":")
		if len(pair) != 2 {
			return nil, fmt.Errorf("invalid field %q", p)
		}
		col := strings.TrimSpace(pair[0])
		typeName := strings.TrimSpace(pair[1])
		mapped, ok := typeMap[typeName]
		if !ok {
			return nil, fmt.Errorf("unsupported type %q", typeName)
		}
		if col == "" {
			return nil, fmt.Errorf("empty column name")
		}
		res = append(res, field{
			Column:  toSnake(col),
			GoName:  toPascal(col),
			SQLType: mapped.sql,
			GoType:  mapped.goT,
		})
	}
	return res, nil
}

func nextMigrationVersion(dir string) (int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, fmt.Errorf("read migrations dir: %w", err)
	}
	versions := make([]int, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		prefix := strings.SplitN(e.Name(), "_", 2)[0]
		v, err := strconv.Atoi(prefix)
		if err != nil {
			continue
		}
		versions = append(versions, v)
	}
	if len(versions) == 0 {
		return 1, nil
	}
	sort.Ints(versions)
	return versions[len(versions)-1] + 1, nil
}

func toSnake(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	var b strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(r + 32)
			continue
		}
		if r == ' ' || r == '-' {
			b.WriteByte('_')
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func toPascal(s string) string {
	s = toSnake(s)
	parts := strings.Split(s, "_")
	for i := range parts {
		if parts[i] == "" {
			continue
		}
		parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
	}
	return strings.Join(parts, "")
}

func exitWithError(err error) {
	fmt.Fprintf(os.Stderr, "entitygen: %v\n", err)
	os.Exit(1)
}
