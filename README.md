# Локальный запуск

Проект состоит из HTTP-калькулятора и генератора запросов. Калькулятор вызывает C-библиотеку для сложения и Rust-библиотеку для вычитания.

Все команды выполняются из корня проекта. Калькулятор и генератор запускаются в отдельных терминалах.

## Windows (PowerShell, x64)

Потребуются:

- Go 1.25.7 или новее.
- GCC для Windows x64 (например, TDM-GCC или MinGW-w64), команда `gcc` должна быть доступна в `PATH`.
- Rust и Cargo с установленной целью `x86_64-pc-windows-gnu`.

Если Rust установлен через rustup, добавьте цель для Windows GNU:

```powershell
rustup target add x86_64-pc-windows-gnu
```

### Сборка

Соберите библиотеки и два приложения:

```powershell
powershell -ExecutionPolicy Bypass -File .\source\build.ps1
go build -o calculator.exe ./cmd/calculator
go build -o generator.exe ./cmd/generator
```

Скрипт создаст `source/libcalculator.dll` и `source/libcalculator_rust.dll`.

### Запуск

В первом терминале запустите калькулятор:

```powershell
.\calculator.exe --host 127.0.0.1 --port 8080 --c-lib ./source/libcalculator.dll --rust-lib ./source/libcalculator_rust.dll --interval 5
```

Во втором терминале запустите генератор:

```powershell
.\generator.exe --url http://127.0.0.1:8080/calc --threads 10 --interval 0.1 --timeout 5
```

## Linux / WSL

Потребуются Go 1.25.7 или новее, GCC, Rust и Cargo. В WSL выполняйте команды внутри Linux-дистрибутива, используя Linux-версию Go.

### Сборка

```bash
bash source/build.sh
go build -o calculator ./cmd/calculator
go build -o generator ./cmd/generator
```

Скрипт создаст `source/libcalculator.so` и `source/libcalculator_rust.so`.

### Запуск

В первом терминале:

```bash
./calculator --host 127.0.0.1 --port 8080 --c-lib ./source/libcalculator.so --rust-lib ./source/libcalculator_rust.so --interval 5
```

Во втором терминале:

```bash
./generator --url http://127.0.0.1:8080/calc --threads 10 --interval 0.1 --timeout 5
```

Windows-калькулятор загружает `.dll`, Linux-калькулятор — `.so`. Приложение и библиотеки должны быть собраны для одной ОС и архитектуры.

## HTTP-запросы

| Метод и URL | Результат |
| --- | --- |
| `GET http://127.0.0.1:8080/health` | Ответ `ok`, если сервер доступен |
| `POST http://127.0.0.1:8080/calc?num=10` | Прибавляет 10 к сумме, вычитает 10 из разности; ответ `ok` |
| `GET http://127.0.0.1:8080/metrics` | Метрики в формате Prometheus |

Пример для PowerShell:

```powershell
Invoke-RestMethod -Method Post -Uri 'http://127.0.0.1:8080/calc?num=10'
Invoke-RestMethod -Uri 'http://127.0.0.1:8080/health'
Invoke-RestMethod -Uri 'http://127.0.0.1:8080/metrics'
```

Пример для Linux:

```bash
curl -X POST 'http://127.0.0.1:8080/calc?num=10'
curl 'http://127.0.0.1:8080/health'
curl 'http://127.0.0.1:8080/metrics'
```

Метрики содержат количество запросов к `/calc` за каждую из последних 60 завершённых секунд и p95/p99 длительности вызовов C и Rust за последние 60 секунд. Длительности указаны в секундах.

## Параметры запуска

### Калькулятор

| Параметр | По умолчанию | Назначение |
| --- | --- | --- |
| `--host` | `0.0.0.0` | Адрес сервера |
| `--port` | `8080` | Порт сервера |
| `--c-lib` | `libcalculator.so` рядом с исполняемым файлом | Путь к C-библиотеке |
| `--rust-lib` | `libcalculator_rust.so` рядом с исполняемым файлом | Путь к Rust-библиотеке |
| `--interval` | `5` | Интервал вывода суммы и разности в секундах, должен быть больше нуля |

В командах выше пути к библиотекам передаются явно. Относительные пути считаются от рабочей директории приложения.

### Генератор

| Параметр | По умолчанию | Назначение |
| --- | --- | --- |
| `--url` | `http://localhost:8080/calc` | Адрес калькулятора |
| `--threads`, `-n` | `10` | Количество параллельных рабочих горутин |
| `--interval` | `0.1` | Пауза после каждого запроса в секундах; `0` — без паузы |
| `--timeout` | `5` | Таймаут HTTP-запроса в секундах |

Генератор отправляет случайные целые числа от −100 до 100. Для запуска без паузы между запросами задайте `--interval 0`.

## Запуск в GoLand на Windows

Создайте две конфигурации **Go Build** с **Run kind: Package**:

| Настройка | Калькулятор | Генератор |
| --- | --- | --- |
| Package path | `PythonToGoMigrationTest/cmd/calculator` | `PythonToGoMigrationTest/cmd/generator` |
| Working directory | Корень проекта | Корень проекта |
| Program arguments | `--host 127.0.0.1 --port 8080 --c-lib ./source/libcalculator.dll --rust-lib ./source/libcalculator_rust.dll --interval 5` | `--url http://127.0.0.1:8080/calc --threads 10 --interval 0.1 --timeout 5` |

Перед запуском соберите DLL командой из раздела Windows. В **Program arguments** указываются только параметры, без имени исполняемого файла. Для отладки используйте **Debug**.

## Остановка

Нажмите `Ctrl+C` в терминале генератора, затем в терминале калькулятора. Генератор завершит работу и выведет статистику через `slog`. Калькулятор выполнит graceful shutdown с таймаутом 10 секунд и выведет итоговые сумму и разность.
