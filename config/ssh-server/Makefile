restart:
	make stop
	make start

start:
	-ssh-add -d ssh-nopassphrase.key
	-ssh-add -d ssh.key
	docker-compose up -d --force-recreate

stop:
	docker-compose stop -t 0
	docker-compose rm --force

shell:
	docker-compose exec openssh-server bash

keys:
	rm -fr ssh.*
	ssh-keygen -t rsa -f ssh-nopassphrase.key -q -N ""
	ssh-keygen -t rsa -f ssh.key -q -N "passphrase"

logs:
	docker-compose logs -f
