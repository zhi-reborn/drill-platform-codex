import { useAuthStore } from '@/stores/auth';
export function useAuth() {
    const store = useAuthStore();
    return {
        user: () => store.user,
        isAuthenticated: () => store.isAuthenticated,
        role: () => store.role,
        hasRole: (r) => store.hasRole(r),
        hasPermission: (p) => store.hasPermission(p),
        login: store.loginWithCredentials,
        loginAsUser: store.loginWithUser,
        logout: store.logout,
    };
}
