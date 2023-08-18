import Head from 'next/head';
import { type ReactNode } from 'react';
import { BASE_URL } from '~/utils/contants';
import Footer from './Footer';
import Nav from './Nav';

export default function Layout({ children }: { children: ReactNode }) {
  return (
    <>
      <Head>
        <title>wakatime.com</title>
        <meta name="description" content="wakatime.com" />
        <link rel="icon" href="/favicon.ico" />
      </Head>
      <div className="flex flex-col items-center">
        <Nav />
        <div className="container flex min-h-screen flex-col items-center py-12">{children}</div>
        <Footer />
      </div>
    </>
  );
}
