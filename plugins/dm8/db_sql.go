package dm8

const (
	QueryDBInstanceRunningInfoSqlStr = `
        SELECT /*+DAMENG_EXPORTER*/
               TO_CHAR(START_TIME,'YYYY-MM-DD HH24:MI:SS'),
               CASE STATUS$ WHEN 'OPEN' THEN '1' WHEN 'MOUNT' THEN '2' WHEN 'SUSPEND' THEN '3' ELSE '4' END AS STATUS,
               CASE MODE$ WHEN 'PRIMARY' THEN '1' WHEN 'NORMAL' THEN '2' WHEN 'STANDBY' THEN '3' ELSE '4' END AS MODE,
               (SELECT COUNT(*) FROM V$TRXWAIT) TRXNUM,
               (SELECT COUNT(*) FROM V$LOCK WHERE BLOCKED=1) DEADLOCKNUM,
               (SELECT COUNT(*) FROM V$THREADS) THREADSNUM,
               DATEDIFF(SQL_TSI_DAY,START_TIME,sysdate) DBSTARTDAY
        
        FROM V$INSTANCE`

	//表空间的使用率
	QueryTablespaceInfoSqlStr = `SELECT /*+DAMENG_EXPORTER*/ F.TABLESPACE_NAME,F.FREE_SPACE   AS "FREE_SIZE",T.TOTAL_SPACE   "TOTAL_SIZE"
FROM (
SELECT TABLESPACE_NAME, ROUND(SUM(BLOCKS * (SELECT PARA_VALUE / 1024 FROM V$DM_INI WHERE PARA_NAME = 'GLOBAL_PAGE_SIZE') / 1024)) FREE_SPACE
        FROM DBA_FREE_SPACE
        GROUP BY TABLESPACE_NAME) F,
      (SELECT TABLESPACE_NAME, ROUND(SUM(BYTES / 1048576)) TOTAL_SPACE
        FROM DBA_DATA_FILES
        GROUP BY TABLESPACE_NAME
        ) T
WHERE F.TABLESPACE_NAME = T.TABLESPACE_NAME`

	//表空间数据文件
	QueryTablespaceFileSqlStr = `SELECT /*+DAMENG_EXPORTER*/ PATH,
            TO_CHAR(TOTAL_SIZE*PAGE) AS TOTAL_SIZE,
            TO_CHAR(FREE_SIZE*PAGE)AS FREE_SIZE,
            AUTO_EXTEND,
            NEXT_SIZE,
            MAX_SIZE
    FROM V$DATAFILE;`

	//查询内存池的状态
	QueryMemoryPoolInfoSqlStr = `SELECT /*+DM_EXPORTER*/ ZONE_TYPE,CURR_VAL,RES_VAL,TOTAL_VAL FROM (
 SELECT 'HJ ZONE' AS ZONE_TYPE,(SELECT SUM(STAT_VAL) FROM V$SYSSTAT WHERE ID IN (114,115)) AS CURR_VAL,(SELECT STAT_VAL FROM V$SYSSTAT WHERE ID IN (145)) AS RES_VAL,(SELECT STAT_VAL FROM V$SYSSTAT WHERE ID IN (144)) AS TOTAL_VAL FROM DUAL UNION ALL
 SELECT 'HAGR ZONE',(SELECT SUM(STAT_VAL) FROM V$SYSSTAT WHERE ID IN (116)),(SELECT STAT_VAL FROM V$SYSSTAT WHERE ID IN (143)),(SELECT STAT_VAL FROM V$SYSSTAT WHERE ID IN (142)) FROM DUAL UNION ALL
 SELECT 'SORT ZONE',(SELECT SUM(STAT_VAL) FROM V$SYSSTAT WHERE ID IN (178)),NULL,(SELECT STAT_VAL FROM V$SYSSTAT WHERE ID IN (177)) FROM DUAL)`

	//查询数据库的会话状态
	QueryDBSessionsStatusSqlStr = `SELECT /*+DM_EXPORTER*/
        DECODE(STATE, NULL, 'TOTAL', STATE) AS STATE_TYPE,
        COUNT(SESS_ID) AS COUNT_VAL
FROM V$SESSIONS
WHERE
        STATE IN ('IDLE', 'ACTIVE')
GROUP BY
        ROLLUP(STATE) union all  select /*+DM_EXPORTER*/ 'MAX_SESSION' STATE_TYPE,para_value from v$dm_ini where para_name = 'MAX_SESSIONS'`

	//查询数据库定时任务错误数量的SQL
	QueryDbJobRunningInfoSqlStr = ` SELECT /*+DM_EXPORTER*/ COUNT(*) error_num FROM (SELECT NAME,ERRINFO FROM SYSJOB.SYSJOBHISTORIES2 WHERE ERRCODE !=0 AND START_TIME >= (SYSDATE-31) AND NAME  IN (SELECT SYSJOBS.NAME
  FROM SYSJOB.SYSJOBSCHEDULES SCHE
  LEFT JOIN SYSJOB.USER_JOBS USERJOB
  ON SCHE.JOBID = USERJOB.JOB LEFT JOIN SYSJOB.SYSJOBSTEPS STEPS
  ON SCHE.JOBID = STEPS.JOBID LEFT JOIN SYSJOB.SYSJOBS SYSJOBS ON SCHE.JOBID = SYSJOBS.ID       
  WHERE  STEPS."TYPE" = 6AND SCHE.VALID = 'Y'
  GROUP BY SYSJOBS.NAME)
  UNION ALL
  SELECT NAME,ERRINFO FROM SYSJOB.SYSSTEPHISTORIES2 WHERE ERRCODE !=0 AND START_TIME  >= (SYSDATE-31) AND NAME IN (SELECT SYSJOBS.NAME
  FROM SYSJOB.SYSJOBSCHEDULES SCHE LEFT JOIN SYSJOB.USER_JOBS USERJOB
  ON SCHE.JOBID = USERJOB.JOB LEFT JOIN SYSJOB.SYSJOBSTEPS STEPS
  ON SCHE.JOBID = STEPS.JOBID LEFT JOIN SYSJOB.SYSJOBS SYSJOBS ON SCHE.JOBID = SYSJOBS.ID       
  WHERE STEPS."TYPE" = 6 AND SCHE.VALID = 'Y' 
  GROUP BY SYSJOBS.NAME))`
	//查询数据库的慢SQL
	QueryDbSlowSqlInfoSqlStr = `select /*+DM_EXPORTER*/ *  from ( SELECT DATEDIFF(MS,LAST_RECV_TIME,SYSDATE) EXEC_TIME,
                            DBMS_LOB.SUBSTR(SF_GET_SESSION_SQL(SESS_ID)) SLOW_SQL,
                            SESS_ID,
                            CURR_SCH,
                            THRD_ID,
                            LAST_RECV_TIME,
                            SUBSTR(CLNT_IP,8,13) CONN_IP
                       FROM V$SESSIONS
                      WHERE  1=1 
                   and STATE='ACTIVE'
                   ORDER BY 1 DESC) 
             where EXEC_TIME >= ? LIMIT ?`
	//查询监视器信息
	QueryMonitorInfoSqlStr = `select /*+DM_EXPORTER*/ * from v$dmmonitor`
	//查询数据库的语句执行次数
	QuerySqlExecuteCountSqlStr = `select /*+DM_EXPORTER*/  NAME,STAT_VAL from v$sysstat where name in ('select statements','insert statements','delete statements','update statements','ddl statements','transaction total count','select statements in pl/sql','insert statements in pl/sql','delete statements in pl/sql','update statements in pl/sql','DDL in pl/sql count','dynamic exec in pl/sql')`
	//查询数据库参数
	QueryParameterInfoSql = `select /*+DM_EXPORTER*/ para_name,para_value from v$dm_ini where para_name in  ( 'MAX_SESSIONS','REDOS_BUF_NUM','REDOS_BUF_SIZE')`
	//查询检查点信息
	QueryCheckPointInfoSql = `select /*+DM_EXPORTER*/ CKPT_TOTAL_COUNT,CKPT_RESERVE_COUNT,CKPT_FLUSHED_PAGES,CKPT_TIME_USED from V$CKPT`
	//查询用户信息
	QueryUserInfoSqlStr = `SELECT 
                       /*+DM_EXPORTER*/ 
                       A.USERNAME ,
                       CASE B.RN_FLAG WHEN '0' THEN 'N' WHEN '1' THEN 'Y' END AS READ_ONLY,
                       CASE A.ACCOUNT_STATUS WHEN 'LOCKED' THEN '锁定' WHEN 'OPEN' THEN '正常' ELSE '异常' END AS ACCOUNT_STATUS,
                       TO_CHAR(A.EXPIRY_DATE,'YYYY-MM-DD HH24:MI:SS') AS EXPIRY_DATE,
                       to_char(round(datediff(DAY,TO_CHAR(sysdate,'YYYY-MM-DD HH24:MI:SS'),TO_CHAR(A.EXPIRY_DATE,'YYYY-MM-DD HH24:MI:SS')),2)) AS EXPIRY_DATE_DAY,
                       A.DEFAULT_TABLESPACE,
                       A.PROFILE,
                       TO_CHAR(A.CREATED,'YYYY-MM-DD HH24:MI:SS') AS CREATE_TIME
                  FROM DBA_USERS A, 
                       SYSUSERS B 
                 WHERE A.USER_ID=B.ID and A.USERNAME NOT IN('SYS','SYSSSO','SYSAUDITOR')`
	//查询数据库授权信息
	QueryDbGrantInfoSql = `SELECT /*+DM_EXPORTER*/ CASE WHEN expired_date IS NULL THEN '' ELSE TO_CHAR(expired_date, 'yyyyMMdd')  END AS expired_date FROM V$LICENSE`
	//查询主备库的同步堆积信息
	QueryStandbyInfoSql = `SELECT /*+DM_EXPORTER*/ task_mem_used, task_num FROM v$rapply_sys`
)
