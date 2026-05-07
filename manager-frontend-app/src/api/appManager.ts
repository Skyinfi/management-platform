import { apiClient } from './client'
import type {
  ApiResponse,
  AppActionResponse,
  ApplicationLogResponse,
  DashboardResponse,
  ListApplicationsResponse,
} from './types'

export const appManagerApi = {
  getDashboard: () => apiClient.get<ApiResponse<DashboardResponse>>('/dashboard'),
  listApplications: () => apiClient.get<ApiResponse<ListApplicationsResponse>>('/applications'),
  startApplication: (name: string) => apiClient.post<ApiResponse<AppActionResponse>>(`/applications/${name}/start`),
  stopApplication: (name: string) => apiClient.post<ApiResponse<AppActionResponse>>(`/applications/${name}/stop`),
  restartApplication: (name: string) => apiClient.post<ApiResponse<AppActionResponse>>(`/applications/${name}/restart`),
  fetchApplicationLogs: (name: string) => apiClient.get<ApiResponse<ApplicationLogResponse>>(`/applications/${name}/logs`),
}
