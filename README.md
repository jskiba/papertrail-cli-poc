# papertrail-cli-poc

# Run
1. [Install Go](https://go.dev/doc/install)
2. Clone the repository
```
git clone https://github.com/jskiba/papertrail-cli-poc.git
```
3. Build
```
go build .
```
4. Run
```
env PAPERTRAIL_TOKEN=<your-token> ./papertrail-cli-poc <lines-of-logs>
```

for better readability:
```
brew install lnav
env PAPERTRAIL_TOKEN=<your-token> ./papertrail-cli-poc <lines-of-logs> | lnav
```
