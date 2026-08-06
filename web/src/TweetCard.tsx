import React from 'react';
import type { TweetResponseTweet } from './types/tweet'

type TweetProps = {
    tweet: TweetResponseTweet
};

const TweetCard: React.FC<TweetProps> = ({ tweet }) => {
    const escapeHtml = (str: string) => {
        return str
            .replace(/&/g, '&amp;')
            .replace(/</g, '&lt;')
            .replace(/>/g, '&gt;')
            .replace(/"/g, '&quot;')
            .replace(/'/g, '&#039;');
    };

    const formatText = (text: string) => {
        let safeText = escapeHtml(text);

        if (tweet.urls) {
            tweet.urls.forEach(u => {
                const escapedUrl = escapeHtml(u.url);
                const escapedExpandedUrl = escapeHtml(u.expanded_url);
                const escapedDisplayUrl = escapeHtml(u.display_url);
                safeText = safeText.replaceAll(
                    escapedUrl,
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
            /https:\/\/t\.co\/[a-zA-Z0-9]+/g,
            (url) => `<a href="${url}" target="_blank" rel="noopener noreferrer" class="text-blue-500 hover:underline">${url}</a>`
        );

        return safeText.replace(/\n/g, '<br />');
    };
    const tweetUrl = `https://x.com/${tweet.screen_name}/status/${tweet.id}`

    return (
        <div className="border-b border-dashed border-black p-4 bg-white space-y-2 last:border-b-0">
            <div className="flex items-center gap-2">
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
                className="text-base leading-relaxed"
                dangerouslySetInnerHTML={{__html: formatText(tweet.text)}}
            />
            {tweet.media && tweet.media.length > 0 && (
                <div className="flex pt-2">
                    {tweet.media.map((url, i) => (
                        <img
                            key={i}
                            src={`${url}:thumb`}
                            alt="media"
                            className="rounded-xl max-h-30 object-cover pr-1"
                        />
                    ))}
                </div>
            )}
            <div className="flex items-center gap-2 text-sm text-gray-500">
                {tweet.retweeted && (
                    <div>🔁 retweeted at</div>
                )}
                {tweet.replied && (
                    <div>↪️ </div>
                )}
                {!tweet.retweeted && (
                    <div>created at</div>
                )}
                <div><a href={tweetUrl} target="_blank">{tweet.created}</a></div>
            </div>
        </div>
    );
};

export default TweetCard;