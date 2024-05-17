local: install
	npm run dev --host

install:
	npm install

upgrade:
	npx @astrojs/upgrade 

obsidian_sync:
	bin/obsidian_sync.sh
