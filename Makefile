user-rpc-dev:
	@make -f deploy/mk/user-rpc.mk release-test

user-api-dev:
	@make -f deploy/mk/user-api.mk release-test

social-rpc-dev:
	@make -f deploy/mk/social-rpc.mk release-test

social-api-dev:
	@make -f deploy/mk/social-api.mk release-test

# 发布版本
release-test: user-rpc-dev user-api-dev social-rpc-dev social-api-dev

# 从docker 拉取并部署
install-server:
	cd ./deploy/script && chmod +x release-test.sh && ./release-test.sh