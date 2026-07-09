# 分布式任务调度平台 - Windows 一键启动脚本（示例）
# 使用前请确保：MySQL/Redis 已运行，etcd 已启动
# 用法（在项目根目录执行）：
#   powershell -ExecutionPolicy Bypass -File scripts/start.ps1

$ErrorActionPreference = "Stop"

Write-Host "=== 1. 初始化数据库（仅首次）===" -ForegroundColor Cyan
$init = Read-Host "是否执行 sql/init.sql 初始化数据库？(y/N)"
if ($init -eq "y" -or $init -eq "Y") {
    mysql -uroot -p < sql/init.sql
}

Write-Host "=== 2. 拉取依赖 ===" -ForegroundColor Cyan
go mod tidy

Write-Host "=== 3. 启动 etcd（如尚未运行）===" -ForegroundColor Cyan
# 若 etcd 已在运行可跳过；默认数据目录为当前目录下的 default.etcd
Start-Process -FilePath "etcd" -ArgumentList "--enable-grpc-gateway" -WindowStyle Normal

Write-Host "=== 4. 启动调度中心 (Admin :8080) ===" -ForegroundColor Cyan
Start-Process -FilePath "go" -ArgumentList "run","./cmd/admin" -WindowStyle Normal

Start-Sleep -Seconds 3

Write-Host "=== 5. 启动执行器 (Executor :9999) ===" -ForegroundColor Cyan
Start-Process -FilePath "go" -ArgumentList "run","./cmd/worker" -WindowStyle Normal

Write-Host "=== 启动完成 ===" -ForegroundColor Green
Write-Host "调度中心: http://127.0.0.1:8080/api/job/page"
Write-Host "执行日志: http://127.0.0.1:8080/api/log/page"
