<?php
// Define your secret token here
define('SECRET_TOKEN', 'secretwebhookkakak');
define('TOKEN', 'v4.publik.yiug');
//define('URLSENDMESSAGE', 'https://api.wa.my.id/api/send/message/text');
// jika domain yang atas di blokir gunakan domain yang bawah
define('URLSENDMESSAGE', 'https://cloud.wa.my.id/api/send/message/text');


// Check if the request method is POST
if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    // Get the token from the request headers
    $headers = apache_request_headers();
    $requestToken = isset($headers['secret']) ? $headers['secret'] : '';
    if ($requestToken==''){
        $requestToken = isset($headers['Secret']) ? $headers['Secret'] : '';
    }

    // Verify the token
    if ($requestToken === SECRET_TOKEN) {
        // Get the JSON payload from the POST request
        $jsonPayload = file_get_contents('php://input');

        // Decode the JSON payload
        $data = json_decode($jsonPayload, true);

        // Check if JSON decoding was successful
        if ($data) {
            // The JSON data.
            $msg = array(
                'to' => $data["phone_number"],
                'isgroup' => false,
                'messages' => 'hai kak ',
            );
            // JSON encode the data.
            $jsonData = json_encode($msg);
            // The URL to send the POST request to.
            $url=URLSENDMESSAGE
            //$url = 'https://api.wa.my.id/api/send/message/text';
            // jika domain yang atas di blokir gunakan domain yang bawah
            //$url = 'https://cloud.wa.my.id/api/send/message/text';
            // Initialize cURL session.
            $ch = curl_init($url);
            // Set cURL options.
            curl_setopt($ch, CURLOPT_POST, true);
            curl_setopt($ch, CURLOPT_POSTFIELDS, $jsonData);
            curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
            curl_setopt($ch, CURLOPT_HTTPHEADER, array(
                'accept: application/json',
                'Content-Type: application/json',
                'token: ' . TOKEN // Add your token header here
            ));
            
            // Execute the POST request.
            $response = curl_exec($ch);
            
            // Check for cURL errors
            if (curl_errno($ch)) {
                $error_msg = curl_error($ch);
            }
            
            // Close cURL session.
            curl_close($ch);
            // If there was an error, handle it here.
            if (isset($error_msg)) {
                // Log or echo the error message.
                echo "cURL Error: " . $error_msg;
            }
            
            // Do something with the response.
            echo $response;
            //echo json_encode($msg);
        } else {
            // JSON was not decoded, handle the error
            http_response_code(400);
            echo json_encode(['status' => 'error', 'message' => 'Invalid JSON']);
        }
    } else {
        // Token is invalid, handle the error
        http_response_code(401);
        echo json_encode(['status' => 'error', 'message' => 'Unauthorized access']);
    }
} else {
    // The request method is not POST, handle the error
    http_response_code(405);
    echo json_encode(['status' => 'error', 'message' => 'Method not allowed']);
}
?>