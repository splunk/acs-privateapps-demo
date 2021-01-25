generate-app-package:
	tar -czvf app-package.tar.gz testapp

inspect-app:
	go run ./src/cmd vet app-package.tar.gz --json-report-file=report.json

install-app:
	go run ./src/cmd install ${STACK_NAME} app-package.tar.gz
