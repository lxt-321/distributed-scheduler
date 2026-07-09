package store

import (
	"database/sql"
	"time"

	"dscheduler/model"
)

const jobCols = `id,job_group,job_cron,job_desc,author,executor_route_strategy,executor_handler,executor_param,executor_block_strategy,executor_timeout,executor_fail_retry_count,trigger_status,trigger_last_time,trigger_next_time,update_time`

// ---------- 任务 JobInfo ----------

// LoadTriggerJobs 加载所有处于运行状态的任务（调度引擎启动时调用）
func LoadTriggerJobs() ([]model.JobInfo, error) {
	rows, err := DB.Query("SELECT "+jobCols+" FROM xxl_job_info WHERE trigger_status=?", 1)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanJobs(rows)
}

// GetJob 按 ID 查询任务
func GetJob(id int64) (model.JobInfo, error) {
	var j model.JobInfo
	err := DB.QueryRow("SELECT "+jobCols+" FROM xxl_job_info WHERE id=?", id).
		Scan(&j.ID, &j.JobGroup, &j.JobCron, &j.JobDesc, &j.Author, &j.ExecutorRouteStrategy,
			&j.ExecutorHandler, &j.ExecutorParam, &j.ExecutorBlockStrategy, &j.ExecutorTimeout,
			&j.ExecutorFailRetryCount, &j.TriggerStatus, &j.TriggerLastTime, &j.TriggerNextTime, &j.UpdateTime)
	return j, err
}

// ListJobs 分页查询任务（简化：固定 LIMIT）
func ListJobs(limit int) ([]model.JobInfo, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := DB.Query("SELECT "+jobCols+" FROM xxl_job_info ORDER BY id DESC LIMIT ?", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanJobs(rows)
}

// InsertJob 新增任务，返回自增 ID
func InsertJob(j model.JobInfo) (int64, error) {
	res, err := DB.Exec(`INSERT INTO xxl_job_info
		(job_group,job_cron,job_desc,author,executor_route_strategy,executor_handler,executor_param,executor_block_strategy,executor_timeout,executor_fail_retry_count,trigger_status,update_time)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,NOW())`,
		j.JobGroup, j.JobCron, j.JobDesc, j.Author, j.ExecutorRouteStrategy, j.ExecutorHandler,
		j.ExecutorParam, j.ExecutorBlockStrategy, j.ExecutorTimeout, j.ExecutorFailRetryCount, j.TriggerStatus)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// UpdateJob 更新任务配置
func UpdateJob(j model.JobInfo) error {
	_, err := DB.Exec(`UPDATE xxl_job_info SET
		job_group=?,job_cron=?,job_desc=?,author=?,executor_route_strategy=?,executor_handler=?,
		executor_param=?,executor_block_strategy=?,executor_timeout=?,executor_fail_retry_count=?,
		trigger_status=?,update_time=NOW() WHERE id=?`,
		j.JobGroup, j.JobCron, j.JobDesc, j.Author, j.ExecutorRouteStrategy, j.ExecutorHandler,
		j.ExecutorParam, j.ExecutorBlockStrategy, j.ExecutorTimeout, j.ExecutorFailRetryCount,
		j.TriggerStatus, j.ID)
	return err
}

// DeleteJob 删除任务
func DeleteJob(id int64) error {
	_, err := DB.Exec("DELETE FROM xxl_job_info WHERE id=?", id)
	return err
}

// SetTriggerStatus 启动/停止任务
func SetTriggerStatus(id int64, status int) error {
	_, err := DB.Exec("UPDATE xxl_job_info SET trigger_status=?,update_time=NOW() WHERE id=?", status, id)
	return err
}

func scanJobs(rows *sql.Rows) ([]model.JobInfo, error) {
	var jobs []model.JobInfo
	for rows.Next() {
		var j model.JobInfo
		if err := rows.Scan(&j.ID, &j.JobGroup, &j.JobCron, &j.JobDesc, &j.Author, &j.ExecutorRouteStrategy,
			&j.ExecutorHandler, &j.ExecutorParam, &j.ExecutorBlockStrategy, &j.ExecutorTimeout,
			&j.ExecutorFailRetryCount, &j.TriggerStatus, &j.TriggerLastTime, &j.TriggerNextTime, &j.UpdateTime); err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
	}
	return jobs, rows.Err()
}

// ---------- 执行器分组 JobGroup ----------

// ListGroups 查询所有执行器分组
func ListGroups() ([]model.JobGroup, error) {
	rows, err := DB.Query("SELECT id,app_name,title,address_type,address_list,update_time FROM xxl_job_group ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var groups []model.JobGroup
	for rows.Next() {
		var g model.JobGroup
		if err := rows.Scan(&g.ID, &g.AppName, &g.Title, &g.AddressType, &g.AddressList, &g.UpdateTime); err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, rows.Err()
}

// InsertGroup 新增执行器分组
func InsertGroup(g model.JobGroup) (int64, error) {
	res, err := DB.Exec("INSERT INTO xxl_job_group(app_name,title,address_type,address_list,update_time) VALUES(?,?,?,?,NOW())",
		g.AppName, g.Title, g.AddressType, g.AddressList)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// UpdateGroup 更新执行器分组
func UpdateGroup(g model.JobGroup) error {
	_, err := DB.Exec("UPDATE xxl_job_group SET app_name=?,title=?,address_type=?,address_list=?,update_time=NOW() WHERE id=?",
		g.AppName, g.Title, g.AddressType, g.AddressList, g.ID)
	return err
}

// DeleteGroup 删除执行器分组
func DeleteGroup(id int64) error {
	_, err := DB.Exec("DELETE FROM xxl_job_group WHERE id=?", id)
	return err
}

// GetAppNameByGroup 根据分组 ID 反查 AppName（用于 etcd 服务发现）
func GetAppNameByGroup(groupID int64) (string, error) {
	var app string
	err := DB.QueryRow("SELECT app_name FROM xxl_job_group WHERE id=?", groupID).Scan(&app)
	return app, err
}

// ---------- 执行日志 JobLog ----------

// CreateLog 触发时写入一条执行日志，返回日志 ID
func CreateLog(job model.JobInfo, addr string) int64 {
	res, err := DB.Exec(`INSERT INTO xxl_job_log
		(job_group,job_id,executor_address,executor_handler,executor_param,executor_fail_retry_count,trigger_time,trigger_code)
		VALUES(?,?,?,?,?,?,?,?)`,
		job.JobGroup, job.ID, addr, job.ExecutorHandler, job.ExecutorParam, job.ExecutorFailRetryCount,
		time.Now().Unix(), 0)
	if err != nil {
		return 0
	}
	id, _ := res.LastInsertId()
	return id
}

// UpdateLogTrigger 更新触发结果
func UpdateLogTrigger(logID int64, code int, msg string) {
	DB.Exec("UPDATE xxl_job_log SET trigger_code=?,trigger_msg=? WHERE id=?", code, msg, logID)
}

// UpdateLogHandle 更新执行结果（执行器回调时调用）
func UpdateLogHandle(logID int64, code int, msg string) {
	DB.Exec("UPDATE xxl_job_log SET handle_time=?,handle_code=?,handle_msg=? WHERE id=?",
		time.Now().Unix(), code, msg, logID)
}

// ListLogs 分页查询执行日志
func ListLogs(limit int) ([]model.JobLog, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := DB.Query(`SELECT id,job_group,job_id,executor_address,executor_handler,executor_param,
		executor_sharding_param,executor_fail_retry_count,trigger_time,trigger_code,trigger_msg,
		handle_time,handle_code,handle_msg FROM xxl_job_log ORDER BY id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var logs []model.JobLog
	for rows.Next() {
		var l model.JobLog
		if err := rows.Scan(&l.ID, &l.JobGroup, &l.JobID, &l.ExecutorAddress, &l.ExecutorHandler, &l.ExecutorParam,
			&l.ExecutorShardingParam, &l.ExecutorFailRetryCount, &l.TriggerTime, &l.TriggerCode, &l.TriggerMsg,
			&l.HandleTime, &l.HandleCode, &l.HandleMsg); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, rows.Err()
}
