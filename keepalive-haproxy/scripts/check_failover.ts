// check_failover.ts
import axios from 'axios';

// The virtual IP address where our service is available
const VIP = 'http://172.20.0.100';
const CHECK_INTERVAL = 1000; // Check every second

async function checkHAProxy() {
    try {
        const response = await axios.get(VIP, { timeout: 2000 });
        const timestamp = new Date().toISOString();
        console.log(`[${timestamp}] ${response.data.trim()}`);
    } catch (error) {
        if (axios.isAxiosError(error)) {
            console.error(`[${new Date().toISOString()}] Error: ${error.message}`);
        } else {
            console.error(`[${new Date().toISOString()}] Unexpected error: ${error}`);
        }
    }
}

// Run the first check immediately
checkHAProxy();

// Then continuously check the HAProxy service
setInterval(checkHAProxy, CHECK_INTERVAL);

console.log('Monitoring HAProxy service. Press Ctrl+C to stop.');
console.log(`Sending requests to Virtual IP: ${VIP}`);