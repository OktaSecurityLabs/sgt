## Getting a certificate from letsencrypt via certbot


* install certbot
  ```commandline
  sudo -H pip install certbot
  ```
* create cert with certonly flag, specifying domain

  ```commandline
  sudo certbot certonly --manual --preferred-challenges=dns -d sgt.exampledomain.com --agree-tos
  ```

  Note: if you get an SSL related error when using certbot, try doing a `pip install -U cryptography`

You must then create a dns entry in route 53 for your domain.  The type is a txt record and the content is the
values given to you by the certbot script.  Put then entry in place, then complete the cert process with certbot

