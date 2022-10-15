# Backup Tools

Easy Incremental Backup Tool.

## Get

### Download release
Download the version of the corresponding operating system in the [release](https://github.com/ArsFy/backup-tools/releases)

### Build

```bash
git clone https://github.com/ArsFy/backup-tools.git

cd backup-tools/sender
go build .

cd ../recipient
go build .
```

-----

## Run

### Sender

```js
{
    "server": "http://127.0.0.1:26543", // Recipient 
    "token": "123456", // Token (consistent with recipient)
    "path": "/xxx/xxx", // Backup path
    "exclude": [
        "/image",
        "/test"
    ]  // Exclude Path (relative position)
}
```

Run

```bash
./sender
```

### Recipient

```js
{
    "port": 26543,     // Server Port
    "token": "123456"  // Token (consistent with sender)
}
```

Run

```bash
./recipient
```