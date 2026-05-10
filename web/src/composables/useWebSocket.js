import { ref, onBeforeUnmount } from 'vue';
import { wsClient } from '@/websocket';
export function useWebSocket() {
    const connectionStatus = ref('disconnected');
    function subscribe(channel, handler) {
        wsClient.subscribe(channel, handler);
    }
    function unsubscribe(channel, handler) {
        wsClient.unsubscribe(channel, handler);
    }
    function init(token) {
        wsClient.connect(token);
        wsClient.subscribe('connection_status', (msg) => {
            connectionStatus.value = msg.payload.status;
        });
    }
    onBeforeUnmount(() => { });
    return { connectionStatus, subscribe, unsubscribe, init };
}
