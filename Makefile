
docker-build:
	cp ./credentials/* .
	docker build --no-cache -t gcr.io/mchirico/sshproxy-action:test -f Dockerfile .

push:
	docker push gcr.io/mchirico/sshproxy-action:test

build:
	go build -v .

run:
	docker run --rm -it --name docker-action -p 3000:3000  gcr.io/mchirico/sshproxy-action:test

stop:
	docker stop docker-action
