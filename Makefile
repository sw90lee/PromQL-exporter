# ImageName
img = imageURL
tag = 85
# chartinstal
version = 0.4.0
chart_dir = chartDir
chart_name = chartName
chart_repo = chartDirRepo

### Build
save:
	sed -i 's/version: .*/version: ${tag}/' ${chart_dir}/Chart.yaml
	sed -i 's/tag: .*/tag: ${tag}/' ${chart_dir}/values.yaml
	docker build -t ${img}:${tag} .
	docker save -o ${chart_name}-${tag}.tar ${img}:${tag}
	helm package ${chart_name}

cc:
	echo "Compiling for every OS and Platform"
	set GOOS=linux GOARCH=arm go build -o bin/main-linux-arm main.go
	set GOOS=linux GOARCH=arm64 go build -o bin/main-linux-arm64 main.go
	set GOOS=freebsd GOARCH=386 go build -o bin/main-freebsd-386 main.go

cclinux:
	echo "Compiling for every OS and Platform"
	GOOS=linux GOARCH=arm go build -o bin/main-linux-arm main.go
	GOOS=linux GOARCH=arm64 go build -o bin/main-linux-arm64 main.go
	GOOS=freebsd GOARCH=386 go build -o bin/main-freebsd-386 main.go
