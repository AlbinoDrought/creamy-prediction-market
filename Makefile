all: ui creamy-prediction-market

.PHONY: ui
ui: 
	cd ui && npm ci && cp node_modules/@ffmpeg/core/dist/esm/* public/ffmpeg/ && npm run build

.PHONY: creamy-prediction-market
creamy-prediction-market:
	git archive HEAD -o ui/dist/source.tar.gz
	go build -o creamy-prediction-market
