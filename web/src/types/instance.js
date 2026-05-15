export const DRILL_STATUS_LABELS = {
    pending: '待启动',
    running: '执行中',
    paused: '已暂停',
    completed: '已完成',
    terminated: '已终止',
};
export const STEP_STATUS_LABELS = {
    pending: '待执行',
    running: '执行中',
    completed: '已完成',
    timeout: '已超时',
    skipped: '已跳过',
    issue: '异常',
};
export function getCompletedSteps(steps) {
    return steps.filter(s => s.status === 'completed').length;
}
export function getTotalSteps(steps) {
    return steps.length || 1;
}
