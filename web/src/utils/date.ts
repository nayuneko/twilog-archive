export const formatDate = (yyyymmdd: string): string => {
    if (!/^\d{8}$/.test(yyyymmdd)) return yyyymmdd;
    const year = yyyymmdd.slice(0, 4);
    const month = yyyymmdd.slice(4, 6);
    const day = yyyymmdd.slice(6, 8);
    const date = new Date(`${year}-${month}-${day}`);

    const weekday = new Intl.DateTimeFormat('ja-JP', { weekday: 'short' }).format(date);

    return `${year}年${month}月${day}日 (${weekday})`;
}
