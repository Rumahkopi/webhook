# PHP Hosting

Untuk yang menggunakan PHP di web hosting bisa memakai contoh kode [berikut](index.php).

## Pengecekan Blacklist IP dari Hosting
Hal yang pertama dilakukan adalah, pastikan hosting anda tidak memblokir domain atau ip dari alamat api.wa.my.id
Cara pengecekan bisa lewat SSH dari hosting, contoh jika hosting memblokir alamat api wa maka akan seperti ini:  
![image](https://github.com/whatsauth/webhook/assets/11188109/6c58afed-d8aa-4fa6-a1d4-a35a27ed7e6c)  
atau bisa juga ketika melakukan test menggunakan postman ke webhook di hosting, akan terasa lama dan muncul keterangan seperti ini:  
![image](https://github.com/whatsauth/webhook/assets/11188109/17676d2d-b1b3-4b54-ad99-4c9637f8b6fa)  

Cara mensiasatinya adalah dengan mengganti alamat domain api.wa.my.id dengan cloud.wa.my.id  
![image](https://github.com/whatsauth/webhook/assets/11188109/8ac887db-c376-4c0b-a1a1-26c3f966b47c)
Agar alamat webhook di hosting juga bisa diakses oleh server whatsauth maka buatlah subdomain yang terdaftar di Cloudflare.
Pada bagian domain klik Create A New Domain  
![image](https://github.com/whatsauth/webhook/assets/11188109/caa18396-006b-4de7-838c-c4400bdafc82)  
Masukkan nama domain atau subdomain yang sudah terdaftar di cloudflare  
![image](https://github.com/whatsauth/webhook/assets/11188109/be2ab7f8-983d-414b-b7e6-513e01c45581)  
Berhasil terlihat Root Dokumen domain sama antara domain hosting dengan domain yang terdaftar cloudflare  
![image](https://github.com/whatsauth/webhook/assets/11188109/00b8e20c-e720-4e1c-9a08-e8705a5a78e5)  
Sekarang masuk ke Dashboard Cloudflare, pilih domain yang tadi di daftarkan masuk ke menu DNS aktifkan awan kuning nya  
![image](https://github.com/whatsauth/webhook/assets/11188109/3a4a32a6-46da-41ae-b470-633218d4f97c)  
Selanjutnya tinggal pakai URL webhook dari nama domain yang ada di cloudflare, dari yang sebelumnya yusrilhelmi.net/webhook menjadi yusril.vas.web.id/webhook.  
![image](https://github.com/whatsauth/webhook/assets/11188109/77748fec-0039-44cf-a46d-a2b2d448cf68)  
Atau Lebih baik langsung kontak saja support penyedia jasa hosting, untuk membuka blokir IP dari dan ke api.wa.my.id

## Langkah langkah
Buka Cpanel dan masuk ke File Manager buat folder baru bernama webhook di dalam public_html  
![image](https://github.com/whatsauth/webhook/assets/11188109/1a39bd75-1f86-4b38-a068-8becc87f087e)  
Kemudian buat file index.php di dalam folder webhook yang tadi dibuat  
![image](https://github.com/whatsauth/webhook/assets/11188109/a90824d4-f75e-4948-97c9-c2f1d6e19780)  
Kemudian klik edit file index.php yang tadi kita buat  
![image](https://github.com/whatsauth/webhook/assets/11188109/d5d348ab-17af-4c31-abbe-9e53ac54d919)  
Paste kan kode yang dicontohkan di atas, edit bagian SECRET_TOKEN dan TOKEN  
![image](https://github.com/whatsauth/webhook/assets/11188109/70022ace-aa63-48fb-9cca-14c42578a402)  
Klik simpan perubahan, kita cek langsung buka URL dari file tersebut tampak seperti ini  
![image](https://github.com/whatsauth/webhook/assets/11188109/3f244557-70db-4a01-8fe7-26644f38b970)  
Mari kita test dengan PostMan atau Thunder Client

