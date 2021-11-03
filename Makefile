build-cloudctl:
	go build -o cloudCtl ./src/cmd

generate-app-package:
	tar zcf app-package.tar.gz testapp

inspect-app:
	./cloudCtl vet app-package.tar.gz --json-report-file=report.json

inspect-app-victoria:
	./cloudCtl vet app-package.tar.gz --json-report-file=report.json --victoria

install-app:
	./cloudCtl install ${STACK_NAME} app-package.tar.gz

install-app-victoria:
	./cloudCtl install ${STACK_NAME} app-package.tar.gz --victoria

uninstall-app:
	./cloudCtl uninstall ${STACK_NAME} testapp

uninstall-app-victoria:
	./cloudCtl uninstall ${STACK_NAME} testapp --victoria
