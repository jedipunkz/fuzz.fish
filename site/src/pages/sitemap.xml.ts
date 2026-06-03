const siteUrl = 'https://jedipunkz.github.io';
const base = import.meta.env.BASE_URL;
const pageUrl = `${siteUrl}${base}`;

export function GET() {
  return new Response(
    `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url>
    <loc>${pageUrl}</loc>
    <changefreq>weekly</changefreq>
    <priority>1.0</priority>
  </url>
</urlset>
`,
    {
      headers: {
        'Content-Type': 'application/xml',
      },
    },
  );
}
