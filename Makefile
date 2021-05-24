build-cloudctl:
	go build -o cloudCtl ./src/cmd

generate-app-package:
	tar zcf app-package.tar.gz testapp

inspect-app:
	./cloudCtl vet app-package.tar.gz --json-report-file=report.json

install-app:
	./cloudCtl install ${STACK_NAME} app-package.tar.gz

uninstall-app:
	./cloudCtl uninstall ${STACK_NAME} testapp

