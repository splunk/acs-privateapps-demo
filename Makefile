generate-app-package:
	tar -czvf app-package.tar.gz test-app

inspect-app:
	go run ./src/cmd vet app-package.tar.gz --json-report-file=report.json



