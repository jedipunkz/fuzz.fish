const siteUrl = 'https://jedipunkz.github.io';
const base = import.meta.env.BASE_URL;

export function GET() {
  return new Response(
    `User-agent: *
Allow: ${base}

Sitemap: ${siteUrl}${base}sitemap.xml
`,
    {
      headers: {
        'Content-Type': 'text/plain',
      },
    },
  );
}
