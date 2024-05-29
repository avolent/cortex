local: install
	npm run dev --host

install:
	npm install

upgrade:
	npx @astrojs/upgrade 

obsidian_sync:
	brew install pyenv
	python3.12 bin/obsidian_sync.py