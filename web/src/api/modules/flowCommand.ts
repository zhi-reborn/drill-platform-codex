import { apiRequest } from '../request'
import type { FlowCommand } from '@/types/flowCommand'

export const flowCommandApi = {
  get: (id: number) => {
    if (!Number.isFinite(id) || id <= 0) {
      return Promise.reject(new Error('无效的命令 ID'))
    }
    return apiRequest<FlowCommand>({
      url: `/v1/flow-commands/${id}`,
      method: 'GET',
    })
  },
}
