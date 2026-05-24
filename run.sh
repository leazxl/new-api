#!/bin/bash
# ============================================================
# New API - 后台服务管理脚本
# ============================================================
# 用法:
#   ./run.sh start          启动服务（后台运行）
#   ./run.sh stop           停止服务
#   ./run.sh restart        重启服务
#   ./run.sh status         查看运行状态
#   ./run.sh logs [行数]     查看日志（默认 50 行，实时跟踪）
#
# 示例:
#   ./run.sh start          首次启动
#   ./run.sh logs 200       查看最近 200 行日志
#   ./run.sh restart        修改 .env 后重启生效
# ============================================================
# 前置条件:
#   1. 当前目录下存在 new-api 可执行文件和 .env 配置文件
#   2. 确保 data/ 目录存在（SQLite 模式）
# ============================================================

set -e

APP_NAME="new-api"
# 脚本所在目录即为应用根目录
APP_DIR="$(cd "$(dirname "$0")" && pwd)"
PID_FILE="$APP_DIR/.new-api.pid"
LOG_FILE="$APP_DIR/logs/new-api.log"

cd "$APP_DIR"
mkdir -p logs

# ---------- 内部函数 ----------

# 判断进程是否在运行
is_running() {
    [ -f "$PID_FILE" ] && kill -0 "$(cat $PID_FILE)" 2>/dev/null
}

# ---------- 对外命令 ----------

start() {
    # 确保可执行权限（Windows 传过来的文件可能没有 +x）
    chmod +x "$APP_NAME"

    if is_running; then
        echo "$APP_NAME is already running (PID: $(cat $PID_FILE))"
        return 1
    fi
    echo -n "Starting $APP_NAME... "
    # nohup 忽略 HUP 信号，& 放入后台，所有输出追加到日志文件
    nohup ./new-api >> "$LOG_FILE" 2>&1 &
    echo $! > "$PID_FILE"
    echo "done (PID: $!)"
}

stop() {
    if ! is_running; then
        echo "$APP_NAME is not running"
        rm -f "$PID_FILE"
        return 0
    fi
    local pid=$(cat "$PID_FILE")
    echo -n "Stopping $APP_NAME (PID: $pid)... "
    # 先发送 TERM 信号优雅退出
    kill "$pid" 2>/dev/null || true
    # 等待进程退出，最多 10 秒
    for i in $(seq 1 10); do
        if ! kill -0 "$pid" 2>/dev/null; then
            break
        fi
        sleep 1
    done
    # 超时后强制终止
    if kill -0 "$pid" 2>/dev/null; then
        echo -n "force killing... "
        kill -9 "$pid" 2>/dev/null || true
        sleep 1
    fi
    rm -f "$PID_FILE"
    echo "stopped"
}

restart() {
    stop
    sleep 1
    start
}

status() {
    if is_running; then
        echo "$APP_NAME is running (PID: $(cat $PID_FILE))"
    else
        echo "$APP_NAME is not running"
    fi
}

logs() {
    local lines="${1:-50}"
    tail -n "$lines" -f "$LOG_FILE"
}

# ---------- 入口 ----------

case "${1:-start}" in
    start)   start ;;
    stop)    stop ;;
    restart) restart ;;
    status)  status ;;
    logs)    logs "${2:-50}" ;;
    *)
        echo "Usage: $0 {start|stop|restart|status|logs [lines]}"
        exit 1
        ;;
esac
