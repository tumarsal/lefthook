# Lefthook: интеграция в Go-пакет (руководство для агента)

Документ описывает, как подключить lefthook как библиотеку в другой Go-модуль и запускать Git hooks программно. Предназначен для AI-агентов и разработчиков, которые встраивают lefthook в CI-раннер, IDE-плагин, custom CLI или тестовый harness.

## Кратко

| Задача | Пакет | API |
|--------|-------|-----|
| Запустить hook «как CLI» | `github.com/evilmartians/lefthook/v2/lefthook` | `lefthook.New(...).Run(ctx, "pre-commit")` |
| Расширить `skip`/`only` | `github.com/evilmartians/lefthook/v2/skip` | `skip.NewSkipChecker(skip.WithCondition(...))` |
| Полный контроль (только внутри репозитория lefthook) | `internal/command`, `internal/run` | см. раздел «Низкий уровень» |

**Важно:** пакеты `internal/*` нельзя импортировать из другого Go-модуля. Для внешней интеграции используйте только `lefthook` и `skip`.

---

## Предварительные условия

1. **Go-модуль** с зависимостью:

```bash
go get github.com/evilmartians/lefthook/v2@latest
```

2. **Git-репозиторий** — lefthook определяет корень через `git rev-parse --show-toplevel`. Запуск из не-git директории завершится ошибкой при `New()`.

3. **Конфигурация** в корне репозитория (хотя бы один файл):
   - `lefthook.yml` / `lefthook.yaml`
   - `.lefthook.yml` / `.lefthook.yaml`
   - `.config/lefthook.yml`
   - локальные: `lefthook-local.yml`, `.lefthook-local.yml`

4. **Hook в конфиге** — имя должно совпадать с ключом в YAML (`pre-commit`, `pre-push`, custom hooks вроде `lint`).

---

## Архитектура

```
Ваш модуль
    │
    ├─ lefthook.New(WithSkipChecker(...))     ← публичный фасад
    │       └─ internal/command.Lefthook
    │               ├─ LoadConfig()           ← читает lefthook.yml
    │               └─ Run(RunArgs)           ← оркестрация
    │                       └─ internal/run.Run
    │                               └─ controller.NewController(WithSkipChecker(...))
    │                                       └─ skip.Checker.Check(skip, only)
    │                                       └─ выполнение jobs
    │
    └─ skip.NewSkipChecker(WithCondition(...))  ← расширяемые условия skip/only
```

Поток выполнения hook:

1. Загрузка и merge конфигов (`lefthook.yml` + `lefthook-local.yml` + `extends`).
2. Поиск hook по имени в `cfg.Hooks`.
3. Конвертация legacy `commands`/`scripts` в `jobs`.
4. Проверка `skip`/`only` на уровне hook (через `skip.Checker`).
5. `setup` инструкции (если заданы).
6. Параллельный или последовательный запуск `jobs`.
7. Для каждого job: `skip`/`only`, tags, glob/files → выполнение команды.
8. Summary и exit code (ошибка, если хотя бы один job failed).

---

## Уровень 1 — рекомендуемый: пакет `lefthook`

Минимальная интеграция. Поведение близко к `lefthook run pre-commit`.

### API

```go
package lefthook

func New(opts ...Option) (*App, error)

func WithSkipChecker(c skip.Checker) Option
func WithVerbose(verbose bool) Option
func WithColors(colors string) Option   // "on" | "off" | "auto" (default)

type App struct { ... }

func (a *App) Run(ctx context.Context, hook string, gitArgs ...string) error
```

### Минимальный пример

```go
package main

import (
    "context"
    "log"
    "os"

    "github.com/evilmartians/lefthook/v2/lefthook"
)

func main() {
    app, err := lefthook.New()
    if err != nil {
        log.Fatal(err)
    }

    if err := app.Run(context.Background(), "pre-commit"); err != nil {
        os.Exit(1)
    }
}
```

### С кастомным SkipChecker

```go
app, err := lefthook.New(
    lefthook.WithSkipChecker(myChecker),
    lefthook.WithVerbose(true),
    lefthook.WithColors("off"),
)
```

### С git hook arguments

Некоторые hooks (например `prepare-commit-msg`) получают аргументы от Git:

```go
err := app.Run(ctx, "prepare-commit-msg", ".git/COMMIT_EDITMSG", "message")
```

### Поведение `Run`

- Возвращает `nil` — hook успешен или пропущен без ошибок.
- Возвращает `error` — job завершился с ошибкой, конфиг не найден (для unknown hook), прерван по Ctrl+C, и т.д.
- Уважает `LEFTHOOK=0` / `LEFTHOOK=false` — немедленный выход без ошибки.
- Автоустанавливает hooks при первом запуске (как CLI), если не задан `NoAutoInstall`.

---

## Уровень 2 — расширение skip/only: пакет `skip`

CLI lefthook по умолчанию понимает в YAML:

| Форма | Пример |
|-------|--------|
| bool | `skip: true` |
| git state | `skip: merge`, `skip: rebase`, `skip: merge-commit` |
| branch | `skip: [{ ref: "main" }]`, glob: `feat/*` |
| shell | `skip: [{ run: "test -n \"$CI\"" }]` |

Ключи `env`, `file`, `bin` и **любые custom-ключи** работают только если зарегистрировать `Condition` в Go.

### Интерфейсы

```go
package skip

type GitState struct {
    Branch string
    State  string // "merge" | "rebase" | "merge-commit" | ""
}

type Checker interface {
    // true = пропустить выполнение (skip или only не выполнен)
    Check(state func() GitState, skip, only any) bool
}

type Condition interface {
    // true = ключ присутствует в item И условие выполнено
    Match(state func() GitState, item map[string]any) bool
}

type Command interface {
    Run(cmd []string, root string, in, out, errOut io.Writer) error
}

type Logger interface {
    Errorf(format string, args ...any)
}
```

### Конструктор

```go
func NewSkipChecker(opts ...Option) Checker

func WithCommand(cmd Command) Option   // для run: условий (shell)
func WithLogger(l Logger) Option
func WithCondition(c Condition) Option
func WithConditions(c ...Condition) Option
```

`NewSkipChecker` **всегда** регистрирует встроенные `ref` и `run`. Дополнительные условия добавляются через `WithCondition`.

### Семантика skip / only

- **`skip`**: если условие совпало → job/hook **пропускается**.
- **`only`**: если условие **не** совпало → job/hook **пропускается** (инверсия).
- **`skip` имеет приоритет над `only`** при конфликте.
- Элементы массива и ключи в одном map — **OR** (достаточно одного совпадения).
- AND — через `run:` shell-команду или несколько custom Condition в одном map (каждый ключ проверяется отдельно, OR между ними).

### Готовые Condition (opt-in)

```go
skip.Env()                    // env: VAR  или  env: VAR=value
skip.File(fs, repoRoot)       // file: path
skip.Bin()                    // bin: executable-name
```

#### Примеры YAML с зарегистрированными Condition

```yaml
pre-commit:
  jobs:
    - name: lint
      only:
        - bin: golangci-lint
        - file: .golangci.yml
      run: golangci-lint run

    - name: skip-in-ci
      skip:
        - env: CI
      run: make lint
```

### Custom Condition

```go
type vpnCondition struct{}

func (vpnCondition) Match(_ func() skip.GitState, item map[string]any) bool {
    iface, ok := item["vpn"].(string)
    if !ok {
        return false
    }
    return isInterfaceUp(iface)
}
```

```yaml
only:
  - vpn: utun0
```

### Полная сборка Checker

```go
import (
    "github.com/spf13/afero"
    "github.com/evilmartians/lefthook/v2/skip"
)

// cmd должен реализовать skip.Command (например os/exec wrapper)
checker := skip.NewSkipChecker(
    skip.WithCommand(myCmd),
    skip.WithLogger(myLogger),          // optional
    skip.WithCondition(skip.Env()),
    skip.WithCondition(skip.File(afero.NewOsFs(), repoRoot)),
    skip.WithCondition(skip.Bin()),
    skip.WithCondition(vpnCondition{}),
)

app, err := lefthook.New(lefthook.WithSkipChecker(checker))
```

### Адаптер для `os/exec`

```go
import (
    "io"
    "os/exec"
)

type execCmd struct{}

func (execCmd) Run(cmd []string, root string, in, out, errOut io.Writer) error {
    c := exec.Command(cmd[0], cmd[1:]...)
    if root != "" {
        c.Dir = root
    }
    c.Stdin = in
    c.Stdout = out
    c.Stderr = errOut
    return c.Run()
}
```

---

## Уровень 3 — низкий уровень (только fork / тот же module path)

Если вы работаете **внутри репозитория lefthook** (fork), доступны `internal/command` и `internal/run`.

### `command.Lefthook`

```go
import "github.com/evilmartians/lefthook/v2/internal/command"

lh, err := command.NewLefthook(verbose, colors,
    command.WithSkipChecker(checker),
)

cfg, err := lh.LoadConfig()

err = lh.Run(ctx, command.RunArgs{
    Hook:            "pre-commit",
    GitArgs:         []string{},
    Force:           false,   // не пропускать jobs при пустых файлах
    SkipLFS:         false,
    NoTTY:           true,    // без spinner
    NoAutoInstall:   true,    // не ставить git hooks
    NoStageFixed:    false,
    AllFiles:        false,
    FilesFromStdin:  false,
    Verbose:         false,
    Exclude:         []string{},       // exclude files из templates
    Files:           []string{},       // override file templates
    RunOnlyJobs:     []string{},       // --job
    RunOnlyTags:     []string{},       // --tag
    RunOnlyCommands: []string{},       // --command (legacy alias)
    FailOnChanges:     nil,            // *bool
    FailOnChangesDiff: nil,            // *bool
})
```

### `run.Run` — прямой запуск hook struct

```go
import (
    "github.com/evilmartians/lefthook/v2/internal/run"
    "github.com/evilmartians/lefthook/v2/internal/run/controller"
)

results, err := run.Run(ctx, hook, repo, exLogger, run.Options{
    SkipChecker:       checker,
    GitArgs:           args.GitArgs,
    ExcludeFiles:      args.Exclude,
    Files:             args.Files,
    RunOnlyJobs:       args.RunOnlyJobs,
    RunOnlyTags:       args.RunOnlyTags,
    SourceDirs:        sourceDirs,
    Templates:         cfg.Templates,
    GlobMatcher:       cfg.GlobMatcher,
    DisableTTY:        true,
    FailOnChanges:     false,
    FailOnChangesDiff: false,
    Force:             false,
    SkipLFS:           false,
    NoStageFixed:      false,
})
```

### `controller.NewController`

```go
ctrl := controller.NewController(repo, exLogger,
    controller.WithSkipChecker(checker),
)
results, err := ctrl.RunHook(ctx, opts, hook)
```

---

## Результаты выполнения

При использовании `run.Run` / `controller.RunHook` возвращается `[]result.Result`:

```go
type Result struct {
    Sub      []Result
    Name     string
    Duration time.Duration
    // text и status — unexported, используйте методы:
}

func (r Result) Success() bool  // job выполнен успешно
func (r Result) Failure() bool  // job упал
func (r Result) Text() string   // fail_text / сообщение об ошибке
```

Job со статусом skip: `!Success() && !Failure()` (пропущен по условию, tags, пустым файлам и т.д.).

Специальная ошибка:

```go
var failOnChanges *run.FailOnChangesError
errors.As(err, &failOnChanges)
```

---

## Переменные окружения

| Переменная | Эффект |
|------------|--------|
| `LEFTHOOK=0` / `false` | Полное отключение |
| `LEFTHOOK_VERBOSE=1` | Debug-логи |
| `LEFTHOOK_EXCLUDE=tag1,tag2` | Добавляет exclude_tags |
| `LEFTHOOK_OUTPUT` | Управление выводом (`meta,success,failure,summary,skips,execution,...`) |
| `CI` | Влияет на `fail_on_changes: ci/non-ci` |
| `NO_COLOR`, `CLICOLOR`, `CLICOLOR_FORCE` | Цвета терминала |

---

## Обработка ошибок (чеклист для агента)

| Ситуация | Поведение | Действие агента |
|----------|-----------|-----------------|
| Hook не в конфиге | `error: hook X doesn't exist` | Проверить имя и `lefthook.yml` |
| Конфиг не найден | `ConfigNotFoundError` | Создать `lefthook.yml` или указать путь |
| Job failed | `Run` → `error` (пустое сообщение) | Смотреть stdout lefthook (execution output) |
| Hook skipped (skip/only) | `nil`, пустой summary | Норма |
| Unknown hook name (known git hook) | warn + `nil` | Hook просто не настроен |
| Ctrl+C | `"Interrupted"` | Отмена |
| `FailOnChangesError` | typed error | Файлы изменены после hook |

---

## Пошаговый алгоритм для агента

### A. Простой запуск hook

1. `go get github.com/evilmartians/lefthook/v2`
2. Убедиться, что в корне git-репо есть `lefthook.yml` с нужным hook.
3. `app, err := lefthook.New()`
4. `err = app.Run(ctx, "pre-commit")`
5. При `err != nil` → exit code 1.

### B. Добавить проверки env / file / bin в skip/only

1. Определить `repoRoot` (обычно `git rev-parse --show-toplevel` или cwd).
2. Собрать `checker := skip.NewSkipChecker(...)` с `Env()`, `File()`, `Bin()`.
3. Передать в `lefthook.New(lefthook.WithSkipChecker(checker))`.
4. Добавить в YAML соответствующие ключи (`env:`, `file:`, `bin:`).

### C. Добавить полностью custom skip-ключ

1. Реализовать `type X struct{}` с методом `Match(...) bool`.
2. `skip.WithCondition(X{})`.
3. Использовать ключ в YAML: `- mykey: value`.

### D. Запуск только части jobs

Через `command.RunArgs` (низкий уровень) или расширить фасад `lefthook`:

```go
lh.Run(ctx, command.RunArgs{
    Hook:        "pre-commit",
    RunOnlyJobs: []string{"golang"},
    RunOnlyTags: []string{"lint"},
})
```

Фасад `lefthook.App` пока не экспортирует эти флаги — для `--job`/`--tag` используйте `internal/command` или добавьте обёртку в своём модуле.

---

## Полный пример интеграции

```go
package myrunner

import (
    "context"
    "fmt"
    "io"
    "os/exec"
    "strings"

    "github.com/spf13/afero"
    "github.com/evilmartians/lefthook/v2/lefthook"
    "github.com/evilmartians/lefthook/v2/skip"
)

type osCommand struct{}

func (osCommand) Run(cmd []string, root string, in, out, errOut io.Writer) error {
    c := exec.Command(cmd[0], cmd[1:]...)
    c.Dir = root
    c.Stdin = in
    c.Stdout = out
    c.Stderr = errOut
    return c.Run()
}

func repoRoot() (string, error) {
    out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
    if err != nil {
        return "", err
    }
    return strings.TrimSpace(string(out)), nil
}

func RunPreCommit(ctx context.Context) error {
    root, err := repoRoot()
    if err != nil {
        return fmt.Errorf("not in git repo: %w", err)
    }

    checker := skip.NewSkipChecker(
        skip.WithCommand(osCommand{}),
        skip.WithCondition(skip.Env()),
        skip.WithCondition(skip.File(afero.NewOsFs(), root)),
        skip.WithCondition(skip.Bin()),
    )

    app, err := lefthook.New(
        lefthook.WithSkipChecker(checker),
        lefthook.WithColors("auto"),
    )
    if err != nil {
        return err
    }

    if err := app.Run(ctx, "pre-commit"); err != nil {
        return err
    }
    return nil
}
```

---

## Ограничения и заметки

- **`skip`/`only` не наследуются** из group в дочерние jobs — задаются на каждый job отдельно.
- **`job.env`** не участвует в skip-проверках — оценивается окружение процесса lefthook.
- **JSON Schema** не описывает custom-ключи skip — это расширение только через Go.
- **CLI без изменений** — стандартный `lefthook run` использует дефолтный Checker (`ref` + `run`).
- Пакет **`lefthook`** — thin wrapper; для новых опций RunArgs расширяйте фасад или используйте `internal/command`.

---

## Связанные файлы в репозитории

| Путь | Назначение |
|------|------------|
| `lefthook/lefthook.go` | Публичный фасад |
| `skip/skip.go` | Интерфейсы Checker, Condition |
| `skip/env.go`, `file.go`, `bin.go` | Opt-in conditions |
| `internal/command/lefthook.go` | NewLefthook, WithSkipChecker |
| `internal/command/run.go` | RunArgs, оркестрация |
| `internal/run/run.go` | run.Run |
| `internal/run/controller/controller.go` | NewController, WithSkipChecker |
| `docs/configuration/skip.md` | Документация skip для YAML |

---

## Быстрая справка импортов

```go
import (
    "github.com/evilmartians/lefthook/v2/lefthook"  // App, New, WithSkipChecker
    "github.com/evilmartians/lefthook/v2/skip"       // Checker, Condition, NewSkipChecker
)
```

Для fork внутри того же module path дополнительно:

```go
import (
    "github.com/evilmartians/lefthook/v2/internal/command"
    "github.com/evilmartians/lefthook/v2/internal/run"
    "github.com/evilmartians/lefthook/v2/internal/run/controller"
)
```
