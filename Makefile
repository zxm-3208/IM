user-rpc-dev:
	@make -f deploy/mk/user-rpc.mk release-test

# 发布版本
release-test: user-rpc-dev

# 从docker 拉取并部署
install-server:
	cd ./deploy/script && chmod +x release-test.sh && ./release-test.sh