#!/bin/sh
# 所有输出重定向到标准错误，保持标准输出干净
echo "启动agent-forge..." >&2
echo "当前目录: $(pwd)" >&2
echo "应用文件权限检查:" >&2
ls -la /app/agent-forge >&2

# 直接执行agent-forge，不产生额外输出
exec /app/agent-forge
