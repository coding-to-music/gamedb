const $homePage = $('#home-page');

if ($homePage.length > 0) {

    // Sales
    $homePage.on('click', '#sales span[data-sort]:not(.badge-success)', function (e) {
        loadSales($(this).attr('data-sort'));
    });

    // Players
    $homePage.on('click', '#players span[data-sort]:not(.badge-success)', function (e) {
        loadPlayers($(this).attr('data-sort'));
    });

    // Panels
    let maxPanelHeight = 0;
    $panels = $('#panels .card');
    $panels.each(function () {
        if ($(this).height() > maxPanelHeight) {
            maxPanelHeight = $(this).height();
        }
    });
    $panels.height(maxPanelHeight);

    loadSales('top-rated');
    loadPlayers('level');

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
                                            '<>': 'div', 'class': 'icon-name', 'html': [
                                                {
                                                    '<>': 'div', 'class': 'icon', 'html': [{'<>': 'img', 'data-lazy': '${icon}', 'alt': '', 'data-lazy-alt': '${name}'}],
                                                },
                                                {
                                                    '<>': 'div', 'class': 'name', 'html': '${name}'
                                                }
                                            ]
                                        }
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
                                        }
                                    ],
                                },
                                {
                                    '<>': 'td', 'html': [
                                        {
                                            '<>': 'a', 'href': '${store_link}', 'target': '_blank', 'rel': 'nofollow', 'html': [
                                                {
                                                    '<>': 'i', 'class': 'fas fa-link',
                                                }
                                            ],
                                        },
                                    ]
                                },
                            ]
                        },
                        {
                            prepend: false,
                        }
                    );

                    observeLazyImages($container.find('img[data-lazy]'));
                    highLightOwnedGames($('#sales'));
                }
            },
        });
    }

    function loadPlayers(sort) {

        $.ajax({
            url: '/home/players/' + sort + '.json',
            dataType: 'json',
            cache: true,
            success: function (data, textStatus, jqXHR) {

                const $allCols = $('#players span[data-sort]');
                $allCols.removeClass('badge-success');
                $allCols.addClass('cursor-pointer');

                const $thisCol = $('#players span[data-sort="' + sort + '"]');
                $thisCol.addClass('badge-success');
                $thisCol.removeClass('cursor-pointer');

                $('#players tbody tr').remove();

                if (isIterable(data)) {

                    const $container = $('#players tbody');

                    const tds = [
                        {
                            '<>': 'td', 'class': 'font-weight-bold', 'html': '${rank}'
                        },
                        {
                            '<>': 'td', 'class': 'img', 'html': [
                                {
                                    '<>': 'div', 'class': 'icon-name', 'html': [
                                        {
                                            '<>': 'div', 'class': 'icon', 'html': [{'<>': 'img', 'data-lazy': '${avatar}', 'alt': '', 'data-lazy-alt': '${name}'}],
                                        },
                                        {
                                            '<>': 'div', 'class': 'name', 'html': '${name}'
                                        }
                                    ]
                                }
                            ]
                        },
                    ];

                    const $change1 = $('#players .change1');
                    const $change2 = $('#players .change2');

                    switch (sort) {
                        case 'level':
                            tds.push({
                                '<>': 'td', 'class': 'img', 'html': '<div class="icon-name"><div class="icon"><div class="${class}"></div></div><div class="name min nowrap">${level}</div></div>',
                            });
                            tds.push({
                                '<>': 'td', 'nowrap': 'nowrap', 'html': "${badges}"
                            });
                            $change1.html('Level');
                            $change2.html('Badges');
                            break;
                        case 'games':
                            tds.push({
                                '<>': 'td', 'nowrap': 'nowrap', 'html': "${games}"
                            });
                            tds.push({
                                '<>': 'td', 'nowrap': 'nowrap', 'html': "${playtime}"
                            });
                            $change1.html('Games');
                            $change2.html('Playtime');
                            break;
                        case 'bans':
                            tds.push({
                                '<>': 'td', 'nowrap': 'nowrap', 'html': "${game_bans}"
                            });
                            tds.push({
                                '<>': 'td', 'nowrap': 'nowrap', 'html': "${vac_bans}"
                            });
                            $change1.html('Game Bans');
                            $change2.html('VAC Bans');
                            break;
                        case 'profile':
                            tds.push({
                                '<>': 'td', 'nowrap': 'nowrap', 'html': "${friends}"
                            });
                            tds.push({
                                '<>': 'td', 'nowrap': 'nowrap', 'html': "${comments}"
                            });
                            $change1.html('Friends');
                            $change2.html('Comments');
                            break;
                    }

                    $container.json2html(
                        data,
                        {
                            '<>': 'tr', 'data-link': '${link}', 'html': tds,
                        },
                        {
                            prepend: false,
                        }
                    );

                    observeLazyImages($container.find('img[data-lazy]'));
                }
            },
        });
    }

    // // Prices
    // $.ajax({
    //     url: '/home/prices.json',
    //     dataType: 'json',
    //     cache: false,
    //     success: function (data, textStatus, jqXHR) {
    //
    //         $('#prices .fa-spin').remove();
    //         $('#prices table').removeClass('d-none');
    //
    //         addPriceRow(data, false);
    //     },
    // });
    //
    // websocketListener('prices', function (e) {
    //
    //     const data = JSON.parse(e.data);
    //
    //     if (data.Data[13] === user.prodCC) { // CC
    //         if (data.Data[12] < 0) { // Drops
    //             if (data.Data[0] > 0) { // Apps
    //                 addPriceRow([{
    //                     "name": data.Data[3],
    //                     "id": data.Data[0],
    //                     "link": data.Data[5],
    //                     "after": data.Data[7],
    //                     "discount": data.Data[15],
    //                     "time": data.Data[11],
    //                     "avatar": data.Data[4],
    //                 }], true);
    //             }
    //         }
    //     }
    // });
    //
    // function addPriceRow(data, addToTop) {
    //
    //     const $container = $('#prices tbody');
    //
    //     $container.json2html(
    //         data,
    //         {
    //             '<>': 'tr', 'data-app-id': '${id}', 'data-link': '${link}', 'html': [
    //                 {
    //                     '<>': 'td', 'class': 'img', 'html': [
    //                         {
    //                             '<>': 'div', 'class': 'icon-name', 'html': [
    //                                 {
    //                                     '<>': 'div', 'class': 'icon', 'html': [{'<>': 'img', 'data-lazy': '${avatar}', 'alt': '', 'data-lazy-alt': '${name}'}]
    //                                 },
    //                                 {
    //                                     '<>': 'div', 'class': 'name', 'html': '${name}'
    //                                 }
    //                             ],
    //                         },
    //                     ]
    //                 },
    //                 {
    //                     '<>': 'td', 'html': '${after}', 'nowrap': 'nowrap'
    //                 },
    //                 {
    //                     '<>': 'td', 'html': '${discount}%', 'nowrap': 'nowrap'
    //                 },
    //                 {
    //                     '<>': 'td', 'nowrap': 'nowrap', 'html': [
    //                         {
    //                             '<>': 'span', 'data-toggle': 'tooltip', 'data-placement': 'left', 'data-livestamp': '${time}',
    //                         }
    //                     ],
    //                 },
    //             ]
    //         },
    //         {
    //             prepend: addToTop,
    //         }
    //     );
    //
    //     $container.find('tr').slice(15).remove();
    //
    //     observeLazyImages($container.find('img[data-lazy]'));
    //     highLightOwnedGames($('#prices'));
    // }
}
