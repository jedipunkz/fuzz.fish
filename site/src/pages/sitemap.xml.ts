import { absoluteUrl } from '../lib/site';

const pageUrl = absoluteUrl();

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
