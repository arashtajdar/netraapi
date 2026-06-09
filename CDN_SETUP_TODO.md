# Cloudflare CDN & Security Setup TODO

When you purchase your domain name and are ready to secure your video streams, follow these steps in order:

## 1. Purchase & Connect a Domain
1. Buy a domain name (e.g., from Cloudflare Registrar, Namecheap, or GoDaddy).
2. Add the domain to your Cloudflare account.
3. Update your domain's Nameservers at your domain registrar to point to the Cloudflare nameservers provided during setup.

## 2. Connect Your Domain to Your R2 Bucket
1. Go to your Cloudflare Dashboard -> **R2**.
2. Click on your `netra-media` bucket.
3. Go to the **Settings** tab.
4. Under **Public Access**, look for **Custom Domains** and click **Connect Domain**.
5. Enter the subdomain you want to use for your videos (e.g., `cdn.yourdomain.com`) and click **Continue** -> **Connect Domain**.
*> Note: Cloudflare will automatically create the DNS records and issue an SSL certificate for you.*

## 3. Attach the Security Worker to Your Domain
1. Go back to the main Cloudflare Dashboard.
2. Click on **Workers & Pages** in the left sidebar.
3. Click on the worker we deployed: **netra-cdn-auth**.
4. Go to the **Domains** tab at the top.
5. Click the blue **+ Add Domain** button.
6. Enter the **exact same domain** you just attached to your R2 bucket (e.g., `cdn.yourdomain.com`).
*> Note: This tells Cloudflare that before anyone can access the R2 bucket on that domain, they must pass through your security script.*

## 4. Finalize Backend Configuration
1. Go to your **Railway Dashboard**.
2. Open your `api` service -> **Variables**.
3. Ensure the following variable is set:
   - `CDN_SECRET_KEY` = `a1b2c3d4e5f60718293a4b5c6d7e8f90`
4. Update your backend settings to use your new custom domain (`https://cdn.yourdomain.com`) instead of the public `r2.dev` URL for all media streaming.

You're done! Your VOD and media content is now fully protected against unauthorized scraping and hotlinking.
