// Base API URL
const API_BASE = '/api';

// Helper function to display results
function displayResult(elementId, data, isError = false) {
    const resultElement = document.getElementById(elementId);
    resultElement.textContent = JSON.stringify(data, null, 2);
    resultElement.className = isError ? 'result error' : 'result success';
}

// Fetch public data (no authentication required)
async function fetchPublicData() {
    try {
        const response = await fetch(`${API_BASE}/public/info`);
        const data = await response.json();
        displayResult('result', data);
    } catch (error) {
        displayResult('result', { error: error.message }, true);
    }
}

// Fetch user data (simulated authenticated endpoint)
async function fetchUserData() {
    try {
        const response = await fetch(`${API_BASE}/users/current`, {
            headers: {
                'Authorization': 'Bearer demo-token-12345'
            }
        });
        const data = await response.json();
        displayResult('result', data);
    } catch (error) {
        displayResult('result', { error: error.message }, true);
    }
}

// Fetch admin data (CORS test - simulates cross-origin request)
async function fetchAdminData() {
    try {
        const response = await fetch(`${API_BASE}/admin/stats`, {
            headers: {
                'Authorization': 'Bearer admin-token-67890'
            }
        });
        const data = await response.json();
        displayResult('result', data);
    } catch (error) {
        displayResult('result', { error: error.message }, true);
    }
}

// Fetch user profile by ID
async function fetchUserProfile() {
    const userId = document.getElementById('userId').value;
    if (!userId) {
        displayResult('userResult', { error: 'Please enter a user ID' }, true);
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/users/${userId}`);
        const data = await response.json();
        displayResult('userResult', data);
    } catch (error) {
        displayResult('userResult', { error: error.message }, true);
    }
}

// Add Enter key support for user ID input
document.addEventListener('DOMContentLoaded', function() {
    const userIdInput = document.getElementById('userId');
    if (userIdInput) {
        userIdInput.addEventListener('keypress', function(event) {
            if (event.key === 'Enter') {
                fetchUserProfile();
            }
        });
    }
});
