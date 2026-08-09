import type { CalendarData } from '../types/calendar';

export const getMinMaxYearMonth = (data: CalendarData) => {
    const years = Object.keys(data).map(Number).sort((a, b) => a - b);
    if (years.length === 0) return null;

    const minYear = years[0];
    const maxYear = years[years.length - 1];

    const minMonths = Object.keys(data[minYear] || {}).map(Number).sort((a, b) => a - b);
    const maxMonths = Object.keys(data[maxYear] || {}).map(Number).sort((a, b) => a - b);

    return {
        minYear,
        minMonth: minMonths[0],
        maxYear,
        maxMonth: maxMonths[maxMonths.length - 1],
    };
};

export const getAllDatesFromCalendar = (data: CalendarData): string[] => {
    const dates: string[] = [];
    for (const year of Object.keys(data)) {
        for (const month of Object.keys(data[year])) {
            const mm = String(month).padStart(2, '0');
            for (const day of data[year][month]) {
                const dd = String(day).padStart(2, '0');
                dates.push(`${year}${mm}${dd}`);
            }
        }
    }
    return dates.sort();
};

export const getAdjacentDates = (calendarData: CalendarData, currentDate: string) => {
    const allDates = getAllDatesFromCalendar(calendarData);
    if (allDates.length === 0) return { prevDate: null, nextDate: null };

    // currentDateより小さい中で最大のものがprevDate
    let prevDate: string | null = null;
    // currentDateより大きい中で最小のものがnextDate
    let nextDate: string | null = null;

    for (const d of allDates) {
        if (d < currentDate) {
            prevDate = d;
        } else if (d > currentDate) {
            nextDate = d;
            break; // allDatesは昇順ソート済みなので最初に見つかった大きめの日付が最小のnextDate
        }
    }

    return { prevDate, nextDate };
};
