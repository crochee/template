#!/bin/bash

unset COLUMNS

PROG_NAME=$0
ACTION=$1
CURRENT_DIR=$(cd $(dirname $0);pwd)
SCRIPT_HOME=$(cd $(dirname $0)/scripts; pwd)
APP_NAME=go_template

usage() {
    echo "Usage: ${PROG_NAME} {start|stop|restart}"
    exit 2
}

log() {
    echo "[$(date +"%Y-%m-%d %H:%M:%S")][${0##*/}:${FUNCNAME[1]}:${BASH_LINENO}] $*"
}

start()
{
    log "INFO: start ${APP_NAME} beginning."
    if [[ -f ${SCRIPT_HOME}/start.sh ]]; then
        su -c "bash chmod +x ${SCRIPT_HOME}/start.sh"
        EXEC_PATH=$(find ${CURRENT_DIR} -type f -executable -name "${APP_NAME}")
        su -c  "bash ${SCRIPT_HOME}/start.sh ${EXEC_PATH}"
        [[ $? -ne 0 ]] && log "ERROR: start ${APP_NAME} failed." && exit 1
    else
        log "INFO: Doing noting about start."
    fi
    log "INFO: start ${APP_NAME} end."
}

stop()
{
    log "INFO: stop ${APP_NAME} beginning."
    if [[ -f ${SCRIPT_HOME}/stop.sh ]]; then
        su -c "bash chmod +x ${SCRIPT_HOME}/stop.sh"
        su -c "bash ${SCRIPT_HOME}/stop.sh ${APP_NAME}"
        [[ $? -ne 0 ]] && log "ERROR: stop ${APP_NAME} failed." && exit 1
    else
        log "INFO: Doing noting about stop."
    fi
    log "INFO: stop ${APP_NAME} end."
}


#check usage
if [[ $# -lt 1 ]]; then
    usage
fi

case "${ACTION}" in
    start)
        start
        ;;
    stop)
        stop
        ;;
    restart)
        stop
        start
        ;;
    *)
        usage
        ;;
esac
