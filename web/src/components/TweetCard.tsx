import React, { useMemo } from 'react';
import type { TweetResponseTweet } from '../types/tweet';
import MediaEmbed from './MediaEmbed';

type TweetProps = {
    tweet: TweetResponseTweet;
};

const escapeHtml = (str: string) => {
    return str
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#039;');
};

const TweetCard: React.FC<TweetProps> = ({ tweet }) => {
    const formattedText = useMemo(() => {
        let safeText = escapeHtml(tweet.text);

        if (tweet.urls) {
            tweet.urls.forEach(u => {
                const escapedUrl = escapeHtml(u.url);
                const escapedExpandedUrl = escapeHtml(u.expanded_url);
                const escapedDisplayUrl = escapeHtml(u.display_url);
                // escapedUrl の直後に付く 「…」 や 「...」 や空白も巻き込んで置換
                const pattern = new RegExp(escapedUrl.replace(/[.*+?^${}()|[\]\\]/g, '\\$&') + '(?:[…\\s]|\\.{3})*', 'g');
                safeText = safeText.replace(
                    pattern,
                    `<a href="${escapedExpandedUrl}" target="_blank" rel="noopener noreferrer" class="text-blue-500 hover:underline">${escapedDisplayUrl}</a>`
                );
            });
        }

        if (tweet.hashtags) {
            tweet.hashtags.forEach(t => {
                const tag = `#${t}`;
                const escapedTag = escapeHtml(tag);
                const url = `https://x.com/search?q=${encodeURIComponent(tag)}`;
                safeText = safeText.replaceAll(
                    escapedTag,
                    `<a href="${url}" target="_blank" rel="noopener noreferrer" class="text-blue-500 hover:underline">${escapedTag}</a>`
                );
            });
        }

        safeText = safeText.replace(
            /https:\/\/t\.co\/[a-zA-Z0-9]+(?:[…\s]|\.{3})*/g,
            (match) => {
                const cleanUrl = match.replace(/(?:[…\s]|\.{3})+$/, '');
                return `<a href="${cleanUrl}" target="_blank" rel="noopener noreferrer" class="text-blue-500 hover:underline">${cleanUrl}</a>`;
            }
        );

        return safeText.replace(/\n/g, '<br />');
    }, [tweet.text, tweet.urls, tweet.hashtags]);

    const tweetUrl = tweet.retweeted
        ? `https://x.com/i/web/status/${tweet.id}`
        : `https://x.com/${tweet.screen_name}/status/${tweet.id}`;

    return (
        <div className="border-b border-dashed border-black p-3 sm:p-4 bg-white space-y-2 last:border-b-0 overflow-hidden">
            <div className="flex flex-wrap items-center gap-x-2 gap-y-0.5">
                {tweet.name ? (
                    <>
                        <span className="font-bold">{tweet.name}</span>
                        <span className="text-sm text-gray-500">@{tweet.screen_name}</span>
                    </>
                ) : (
                    <span className="font-bold">@{tweet.screen_name}</span>
                )}
            </div>
            <div
                className="text-sm sm:text-base leading-relaxed break-words overflow-wrap-anywhere"
                dangerouslySetInnerHTML={{ __html: formattedText }}
            />
            {tweet.media && tweet.media.length > 0 && (
                <div className="flex pt-2 gap-1.5 overflow-x-auto max-w-full pb-1">
                    {tweet.media.map((url) => (
                        <img
                            key={url}
                            src={`${url}:thumb`}
                            alt="media"
                            className="rounded-xl max-h-24 sm:max-h-32 object-cover shrink-0"
                        />
                    ))}
                </div>
            )}
            <MediaEmbed urls={tweet.urls} />
            <div className="flex flex-wrap items-center gap-x-2 gap-y-1 text-xs sm:text-sm text-gray-500 pt-1">
                {tweet.retweeted && (
                    <div>🔁 retweeted at</div>
                )}
                {tweet.replied && (
                    <div>↪️ </div>
                )}
                {!tweet.retweeted && (
                    <div>created at</div>
                )}
                <div>
                    <a href={tweetUrl} target="_blank" rel="noopener noreferrer" className="hover:underline">
                        {tweet.created}
                    </a>
                </div>
            </div>
        </div>
    );
};

export default TweetCard;