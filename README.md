# shell-checker


#### Install

```bash
 git clone https://github.com/0xWhoknows/shell-checker && cd shell-checker
```

#### Usage

```bash
go run file.go -t <site-title> -f <shell-file> [-r <retry-count>] [-ua <user-agent>]
```

#### Options

- `-t`: Shell Title
- `-f`: File Name
- `-r`: Retry if site down (default 1)
- `-ua` : Add User-agent (it also have default user-agent)

#### Example

```bash
go run check.go -f file.txt
```
