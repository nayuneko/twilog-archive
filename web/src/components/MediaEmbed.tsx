import React from 'react';
import type { Urls } from '../types/tweet';

type Props = {
    urls?: Urls[];
};

export const MediaEmbed: React.FC<Props> = ({ urls }) => {
    if (!urls || urls.length === 0) return null;

    return (
        <div className="mt-2.5 space-y-2">
            {urls.map((u, idx) => {
                const targetUrl = u.expanded_url || u.url;
                if (!targetUrl) return null;

                // 1. YouTube & YouTube Music
                const ytMatch = targetUrl.match(/(?:youtube\.com\/(?:watch\?v=|embed\/)|youtu\.be\/|music\.youtube\.com\/watch\?v=)([a-zA-Z0-9_-]{11})/);
                if (ytMatch && ytMatch[1]) {
                    const videoId = ytMatch[1];
                    const isMusic = targetUrl.includes('music.youtube.com');
                    return (
                        <div key={idx} className="rounded-lg overflow-hidden border border-gray-200 shadow-xs bg-black space-y-1">
                            {isMusic && (
                                <div className="px-3 py-1 bg-red-900/80 text-white text-[11px] font-medium flex items-center gap-1.5">
                                    🎵 YouTube Music
                                </div>
                            )}
                            <div className="relative w-full aspect-video">
                                <iframe
                                    src={`https://www.youtube-nocookie.com/embed/${videoId}`}
                                    title="YouTube video player"
                                    className="absolute top-0 left-0 w-full h-full border-0"
                                    allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
                                    allowFullScreen
                                    loading="lazy"
                                />
                            </div>
                        </div>
                    );
                }

                // 2. Spotify
                const spotifyMatch = targetUrl.match(/open\.spotify\.com\/(track|album|playlist|episode)\/([a-zA-Z0-9]+)/);
                if (spotifyMatch) {
                    const [, type, id] = spotifyMatch;
                    return (
                        <div key={idx} className="rounded-lg overflow-hidden border border-gray-200 shadow-xs bg-black">
                            <iframe
                                src={`https://open.spotify.com/embed/${type}/${id}`}
                                title="Spotify Player"
                                width="100%"
                                height="152"
                                className="border-0"
                                allow="autoplay; clipboard-write; encrypted-media; fullscreen; picture-in-picture"
                                loading="lazy"
                            />
                        </div>
                    );
                }

                // 3. ニコニコ動画
                const nicoMatch = targetUrl.match(/nicovideo\.jp\/watch\/(sm[0-9]+|[0-9]+)/);
                if (nicoMatch) {
                    const videoId = nicoMatch[1];
                    return (
                        <div key={idx} className="relative w-full aspect-video rounded-lg overflow-hidden border border-gray-200 shadow-xs bg-black">
                            <iframe
                                src={`https://embed.nicovideo.jp/watch/${videoId}`}
                                title="Niconico Player"
                                className="absolute top-0 left-0 w-full h-full border-0"
                                allowFullScreen
                                loading="lazy"
                            />
                        </div>
                    );
                }

                // 4. Amazon (Amazon Music または商品)
                let domain = '';
                try {
                    domain = new URL(targetUrl).hostname.replace(/^www\./, '');
                } catch {
                    domain = u.display_url || targetUrl;
                }

                const isAmazon = domain.includes('amazon') || domain.includes('amzn');
                const isAmazonMusic = domain.includes('music.amazon');

                return (
                    <a
                        key={idx}
                        href={targetUrl}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="block rounded-lg border border-gray-200/90 bg-gray-50/80 p-3 hover:bg-gray-100 transition-all shadow-2xs group"
                    >
                        <div className="flex items-center justify-between gap-2">
                            <div className="flex items-center gap-2.5 min-w-0">
                                <span className="text-lg leading-none select-none">
                                    {isAmazonMusic ? '🎧' : isAmazon ? '📦' : '🔗'}
                                </span>
                                <div className="truncate">
                                    <div className="text-xs font-semibold text-gray-800 group-hover:text-blue-600 truncate transition-colors">
                                        {u.display_url || domain}
                                    </div>
                                    <div className="text-[11px] text-gray-500 truncate">
                                        {domain}
                                    </div>
                                </div>
                            </div>
                            <span className="text-xs text-gray-400 group-hover:translate-x-0.5 group-hover:text-blue-500 transition-all">
                                ↗
                            </span>
                        </div>
                    </a>
                );
            })}
        </div>
    );
};

export default MediaEmbed;
