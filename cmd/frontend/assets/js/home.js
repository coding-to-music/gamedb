const $homePage = $('#home-page');

if ($homePage.length > 0) {

    // Sales
    $homePage.on('click', '#sales span[data-sort]:not(.badge-success)', function (e) {
        loadSales($(this).attr('data-sort'));
    });

    // Players
    $homePage.on('click', '#top-players span[data-sort]:not(.badge-success)', function (e) {
        homeTopPlayers($(this).attr('data-sort'));
    });

    // Fix top panel heights
    let maxPanelHeight = 0;
    const $panels = $('#panels .card');
    $panels.each(function () {
        if ($(this).height() > maxPanelHeight) {
            maxPanelHeight = $(this).height();
        }
    });
    $panels.css('min-height', maxPanelHeight + 'px');

    // Load AJAX
    loadAjaxOnObserve({
        // "sales": function () {
        //     loadSales('top-rated');
        // },
        'top-players': function () {
            homeTopPlayers('level');
        },
        'new-players': homeNewPlayers,
        'upcoming-games': homeUpcomingGames,
        'new-releases': homeNewReleases,
        'news': loadNewsSection,
        'tweets': loadHomeTweets,
    });

    function loadHomeTweets() {

        $.ajax({
            url: '/home/tweets.json',
            dataType: 'json',
            cache: true,
            success: function (data, textStatus, jqXHR) {

                if (isIterable(data)) {

                    const $container = $('#tweets .card-body').empty();

                    $container.json2html(
                        data,
                        {
                            '<>': 'div', 'class': 'tweet', 'html': [
                                {
                                    '<>': 'img', 'src': '${image}', 'alt': '${text}',
                                },
                                {
                                    '<>': 'p', 'html': '<i class="fas fa-star text-warning"></i> ${text}',
                                },
                            ],
                        },
                        {
                            prepend: false,
                        },
                    );
                }
            },
        });
    }

    function loadNewsSection() {

        const $news = $('#news .card-body');

        $.ajax({
            type: 'GET',
            url: '/home/news.html',
            dataType: 'html',
            success: function (data, textStatus, jqXHR) {
                if (data) {
                    $news.html(data);
                    observeLazyImages($news.find('img[data-lazy]'));
                }
            },
        });
    }

    function homeNewPlayers() {

        let lastPlayerId = 0;
        const $tbody = $('#new-players tbody');
        const schema = function (fade, instantTime) {
            return {
                '<>': 'tr', 'class': (fade ? 'fade-green' : ''), 'data-app-id': '${id}', 'data-link': '${link}', 'html': [
                    {
                        '<>': 'td', 'class': 'img', 'html': [
                            {
                                '<>': 'a', 'href': '${link}', 'class': 'icon-name', 'html': [
                                    {
                                        '<>': 'div', 'class': 'icon', 'html': [{'<>': 'img', 'data-lazy': '${avatar}', 'alt': '', 'data-lazy-alt': '${name}'}],
                                    },
                                    {
                                        '<>': 'div', 'class': 'name', 'html': '${name}',
                                    },
                                ],
                            },
                        ],
                    },
                    {
                        '<>': 'td', 'nowrap': 'nowrap', 'html': [
                            {
                                '<>': 'span', 'data-livestamp': '${created_at}', 'text': (instantTime ? 'a few seconds ago' : ''),
                            },
                        ],
                    },
                    {
                        '<>': 'td', 'html': [
                            {
                                '<>': 'a', 'href': '${community_link}', 'target': '_blank', 'rel': 'noopener', 'html': [
                                    {'<>': 'i', 'class': 'fas fa-link'},
                                ],
                            },
                        ],
                    },
                ],
            };
        };

        $.ajax({
            url: '/home/new-players.json',
            dataType: 'json',
            cache: true,
            success: function (data, textStatus, jqXHR) {

                if (isIterable(data)) {
                    $tbody.find('tr').remove();
                    $tbody.json2html(data, schema(false, false), {prepend: false});
                    observeLazyImages($tbody.find('img[data-lazy]'));
                }

                lastPlayerId = data[0]['id'];
            },
        });

        websocketListener('profile', function (e) {
            const data = JSON.parse(e.data);
            if (data.Data['queue'] === 'player' && data.Data['id'] !== lastPlayerId && data.Data['new']) {
                lastPlayerId = data.Data['id'];
                $tbody.json2html([data.Data], schema(true, true), {prepend: true});
                $tbody.find('tr').slice(10).remove();
                observeLazyImages($tbody.find('img[data-lazy]'));
            }
        });
    }

    //
    function loadSales(sort) {

        $.ajax({
            url: '/home/sales/' + sort + '.json',
            dataType: 'json',
            cache: true,
            success: function (data, textStatus, jqXHR) {

                const $allCols = $('#sales span[data-sort]');
                $allCols.removeClass('badge-success');
                $allCols.addClass('cursor-pointer');

                const $thisCol = $('#sales span[data-sort="' + sort + '"]');
                $thisCol.addClass('badge-success');
                $thisCol.removeClass('cursor-pointer');

                $('#sales tbody tr').remove();
                $('#sales .change').html(sort);

                if (isIterable(data)) {

                    const $container = $('#sales tbody');

                    $container.json2html(
                        data,
                        {
                            '<>': 'tr', 'data-app-id': '${id}', 'data-link': '${link}', 'html': [
                                {
                                    '<>': 'td', 'class': 'img', 'html': [
                                        {
                                            '<>': 'a', 'href': '${link}', 'class': 'icon-name', 'html': [
                                                {
                                                    '<>': 'div', 'class': 'icon', 'html': [{'<>': 'img', 'data-lazy': '${icon}', 'alt': '', 'data-lazy-alt': '${name}'}],
                                                },
                                                {
                                                    '<>': 'div', 'class': 'name', 'html': '${name}',
                                                },
                                            ],
                                        },
                                    ],
                                },
                                {
                                    '<>': 'td', 'html': '${price}', 'class': 'nowrap',
                                },
                                {
                                    '<>': 'td', 'html': '${rating}',
                                },
                                {
                                    '<>': 'td', 'nowrap': 'nowrap', 'class': 'nowrap', 'html': [
                                        {
                                            '<>': 'span', 'data-toggle': 'tooltip', 'data-placement': 'left', 'data-livestamp': '${ends}',
                                        },
                                    ],
                                },
                                {
                                    '<>': 'td', 'html': [
                                        {
                                            '<>': 'a', 'href': '${store_link}', 'target': '_blank', 'rel': 'noopener', 'html': [
                                                {
                                                    '<>': 'i', 'class': 'fas fa-link',
                                                },
                                            ],
                                        },
                                    ],
                                },
                            ],
                        },
                        {
                            prepend: false,
                        },
                    );

                    observeLazyImages($container.find('img[data-lazy]'));
                    highLightOwnedGames($('#sales'));
                }
            },
        });
    }

    function homeTopPlayers(sort) {

        $.ajax({
            url: '/home/top-players/' + sort + '.json',
            dataType: 'json',
            cache: true,
            success: function (data, textStatus, jqXHR) {

                const $allCols = $('#top-players span[data-sort]');
                $allCols.removeClass('badge-success');
                $allCols.addClass('cursor-pointer');

                const $thisCol = $('#top-players span[data-sort="' + sort + '"]');
                $thisCol.addClass('badge-success');
                $thisCol.removeClass('cursor-pointer');

                $('#top-players tbody tr').remove();

                if (isIterable(data)) {

                    const $container = $('#top-players tbody');

                    const tds = [
                        {
                            '<>': 'td', 'class': 'font-weight-bold', 'html': '${rank}',
                        },
                        {
                            '<>': 'td', 'class': 'img', 'html': [
                                {
                                    '<>': 'a', 'href': '${link}', 'class': 'icon-name', 'html': [
                                        {
                                            '<>': 'div', 'class': 'icon', 'html': [{'<>': 'img', 'data-lazy': '${avatar}', 'alt': '', 'data-lazy-alt': '${name}'}],
                                        },
                                        {
                                            '<>': 'div', 'class': 'name', 'html': '${name}',
                                        },
                                    ],
                                },
                            ],
                        },
                    ];

                    const $change1 = $('#top-players .change1');
                    const $change2 = $('#top-players .change2');

                    switch (sort) {
                        case 'level':
                            tds.push({
                                '<>': 'td', 'class': 'img', 'html': '<div class="icon-name"><div class="icon"><div class="${class}"></div></div><div class="name min nowrap">${level}</div></div>',
                            });
                            tds.push({
                                '<>': 'td', 'nowrap': 'nowrap', 'html': '${badges}',
                            });
                            $change1.html('Level');
                            $change2.html('Badges');
                            break;
                        case 'games':
                            tds.push({
                                '<>': 'td', 'nowrap': 'nowrap', 'html': '${games}',
                            });
                            tds.push({
                                '<>': 'td', 'nowrap': 'nowrap', 'html': '${playtime}',
                            });
                            $change1.html('Games');
                            $change2.html('Playtime');
                            break;
                        case 'bans':
                            tds.push({
                                '<>': 'td', 'nowrap': 'nowrap', 'html': '${game_bans}',
                            });
                            tds.push({
                                '<>': 'td', 'nowrap': 'nowrap', 'html': '${vac_bans}',
                            });
                            $change1.html('Game Bans');
                            $change2.html('VAC Bans');
                            break;
                        case 'profile':
                            tds.push({
                                '<>': 'td', 'nowrap': 'nowrap', 'html': '${friends}',
                            });
                            tds.push({
                                '<>': 'td', 'nowrap': 'nowrap', 'html': '${comments}',
                            });
                            $change1.html('Friends');
                            $change2.html('Comments');
                            break;
                        case 'awards':
                            tds.push({
                                '<>': 'td', 'nowrap': 'nowrap', 'html': '${awards_sent}',
                            });
                            tds.push({
                                '<>': 'td', 'nowrap': 'nowrap', 'html': '${awards_received}',
                            });
                            $change1.html('Given');
                            $change2.html('Received');
                            break;
                    }

                    $container.json2html(
                        data,
                        {
                            '<>': 'tr', 'data-link': '${link}', 'html': tds,
                        },
                        {
                            prepend: false,
                        },
                    );

                    observeLazyImages($container.find('img[data-lazy]'));
                }
            },
        });
    }

    const homeChartsBase = $.extend(true, {}, defaultChartOptions, {
        yAxis: {
            allowDecimals: false,
            title: {
                text: '',
            },
            labels: {
                formatter: function () {
                    return this.value.toLocaleString();
                },
            },
            gridLineColor: darkMode ? '#00ff0011' : '#44774411',
        },
        plotOptions: {
            series: {
                cursor: 'pointer',
            },
        },
        legend: {
            enabled: false,
        },
    });

    function homeNewReleases() {

        $.ajax({
            type: 'GET',
            url: '/home/new-releases.json',
            dataType: 'json',
            cache: true,
            success: function (data, textStatus, jqXHR) {

                let series = [];

                for (const datum of data['players']) {

                    const app = data.apps[datum.key];

                    series.push({
                        app: datum.key,
                        name: app.name,
                        data: datum['value']['max_moving_average'], // max_player_count
                        connectNulls: true,
                    });
                }

                const hc = Highcharts.chart('new-releases', $.extend(true, {}, homeChartsBase, {
                    plotOptions: {
                        series: {
                            point: {
                                events: {
                                    click: function () {
                                        location.href = data.apps[this.series.userOptions.id].path;
                                    },
                                },
                            },
                        },
                    },
                    tooltip: {
                        formatter: function () {
                            return this.series.name + ' had ' + this.y.toLocaleString()
                                + ' players on ' + moment(this.key).format('dddd DD MMM YYYY @ HH:00');
                        },
                    },
                    series: series,
                }));

                const $legend = $('#new-releases-legend');

                $.each(hc.series, function (i, v) {

                    const app = data.apps[v.userOptions.app];

                    $legend.append('<div class="cursor-pointer" style="color:' + v.color + '" data-link="' + app.path + '">'
                        + '<img class="rounded" src="' + app.icon + '" alt="' + v.name + '"> ' + v.name + '</div>');
                });

                homeAttachLegendEvents(hc, $legend, data);
            },
        });
    }

    function homeUpcomingGames() {

        $.ajax({
            type: 'GET',
            url: '/home/upcoming-games.json',
            dataType: 'json',
            cache: true,
            success: function (data, textStatus, jqXHR) {

                let series = [];

                for (const datum of data['followers']) {

                    const app = data.apps[datum.key];

                    series.push({
                        group: app.group,
                        name: app.name,
                        data: datum['value']['max_members_count'],
                        connectNulls: true,
                        date: app.date,
                    });
                }

                const hc = Highcharts.chart('upcoming-games', $.extend(true, {}, homeChartsBase, {
                    plotOptions: {
                        series: {
                            point: {
                                events: {
                                    click: function () {
                                        location.href = data.apps[this.series.userOptions.group].path;
                                    },
                                },
                            },
                        },
                    },
                    tooltip: {
                        formatter: function () {
                            return this.series.name + ' had ' + this.y.toLocaleString()
                                + ' followers on ' + moment(this.key).format('dddd DD MMM');
                        },
                    },
                    series: series,
                }));

                const $legend = $('#upcoming-games-legend');

                $.each(hc.series, function (j, series) {

                    const app = data.apps[series.userOptions.group];

                    $legend.append('<div class="cursor-pointer" style="color:' + series.color + '" data-link="' + app.path + '" data-toggle="tooltip" data-placement="top" title="' + app.date + '">'
                        + '<img class="rounded" src="' + app.icon + '" alt="' + series.name + '"> ' + series.name + '</div>');
                });

                homeAttachLegendEvents(hc, $legend, null);
            },
        });
    }

    function homeAttachLegendEvents(hc, $legend, data) {

        $legend.on('mouseenter', 'div', function () {

            if (data) {
                let max = 0;
                data['players'][$(this).index()]['value']['max_player_count'].forEach(function (item) {
                    max = Math.max(max, item[1]);
                });
                hc.yAxis[0].update({max: max});
            }

            $.each(hc.series, function (j, series) {
                hc.series[j].setState('inactive');
            });

            hc.series[$(this).index()].setState('hover');

        }).on('mouseleave', 'div', function () {

            if (data) {
                hc.yAxis[0].update({max: null});
            }

            $.each(hc.series, function (j, series) {
                hc.series[j].setState('normal');
            });
        });
    }
}
