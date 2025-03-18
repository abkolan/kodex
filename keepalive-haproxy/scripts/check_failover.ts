// check_failover.ts
import axios from 'axios';

// Instead of using the VIP directly, we'll check both mapped ports
const NODE1_PORT = 8001;
const NODE2_PORT = 8002;
const CHECK_INTERVAL = 1000; // Check every second
const TIMEOUT = 1000; // Shorter timeout for quicker checks

async function checkNode(port: number, nodeName: string): Promise<boolean> {
    try {
        const response = await axios.get(`http://localhost:${port}`, { timeout: TIMEOUT });
        const timestamp = new Date().toISOString();
        console.log(`[${timestamp}] ${response.data.trim()}`);
        return true; // Successfully reached this node
    } catch (error) {
        // This node is not responding - this is normal for the inactive node
        return false;
    }
}

async function checkHAProxyNodes() {
    const timestamp = new Date().toISOString();
    
    try {
        // Try node1 first
        const node1Active = await checkNode(NODE1_PORT, "Node1");
        
        // If node1 is not active, try node2
        if (!node1Active) {
            const node2Active = await checkNode(NODE2_PORT, "Node2");
            
            // If neither node responds, log an error
            if (!node2Active) {
                console.error(`[${timestamp}] Error: Neither node is responding. The service might be down.`);
            }
        }
    } catch (error) {
        console.error(`[${timestamp}] Unexpected error: ${error}`);
    }
}

// Run the first check immediately
checkHAProxyNodes();

// Then continuously check the HAProxy nodes
setInterval(checkHAProxyNodes, CHECK_INTERVAL);

console.log('Monitoring HAProxy high availability setup. Press Ctrl+C to stop.');
console.log(`Checking Node1 (port ${NODE1_PORT}) and Node2 (port ${NODE2_PORT})`);