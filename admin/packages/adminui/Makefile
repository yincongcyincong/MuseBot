# 项目目录
PROJECT_DIR := $(CURDIR)
DIST_DIR := $(PROJECT_DIR)/dist

# 默认目标
.PHONY: help
help:
	@echo "用法：make <目标>"
	@echo "常用目标："
	@echo "  dev     - 启动开发服务器 (npm run dev)"
	@echo "  build   - 构建生产版本 (npm run build)"
	@echo "  preview - 预览构建结果 (npm run preview)"
	@echo "  clean   - 删除 dist 目录"
	@echo "  install - 安装依赖"

# 启动开发服务器
.PHONY: dev
dev:
	npm run dev

# 构建生产版本
.PHONY: build
build:
	npm run build
	rm -rf ../../adminui/*
	mv ./dist/* ../../adminui

# 本地预览生产版本
.PHONY: preview
preview:
	npm run preview

# 安装依赖
.PHONY: install
install:
	npm install

# 删除构建结果
.PHONY: clean
clean:
	rm -rf $(DIST_DIR)
